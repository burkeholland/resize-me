//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	hotkeyID          = 0x524d
	quickPickHotkeyID = 0x5251

	wmDestroy       = 0x0002
	wmCommand       = 0x0111
	wmHotkey        = 0x0312
	wmUser          = 0x0400
	wmTrayIcon      = wmUser + 1
	wmApplyHotkey   = wmUser + 2 // dispatches hotkey registration to the message-loop thread
	wmShowMenu      = wmUser + 3 // dispatches tray menu display to the message-loop thread
	wmShowQuickPick = wmUser + 4 // dispatches quick-pick menu display to the message-loop thread
	wmLButtonUp     = 0x0202
	wmRButtonUp     = 0x0205
	wmContextMenu   = 0x007B
	gwHwndNext      = 2

	modAlt      = 0x0001
	modControl  = 0x0002
	modShift    = 0x0004
	modWin      = 0x0008
	modNoRepeat = 0x4000

	nimAdd    = 0x00000000
	nimModify = 0x00000001
	nimDelete = 0x00000002

	nifMessage = 0x00000001
	nifIcon    = 0x00000002
	nifTip     = 0x00000004
	nifInfo    = 0x00000010

	niifNone    = 0x00000000
	niifWarning = 0x00000002

	mfString    = 0x00000000
	mfSeparator = 0x00000800
	mfChecked   = 0x00000008
	mfDisabled  = 0x00000002

	tpmRightButton = 0x0002
	tpmReturnCmd   = 0x0100
	tpmNonotify    = 0x0080

	swpNoZOrder   = 0x0004
	swpNoActivate = 0x0010

	swRestore = 9

	monitorDefaultToNearest = 0x00000002
	dwmwaCloaked            = 14

	cmdPresetBase    = 1000
	cmdCenter        = 2000
	cmdResize        = 2001
	cmdSettings      = 2002
	cmdQuit          = 2003
	cmdAbout         = 2004
	cmdQuickPickBase = 3000

	resizeTolerance                = 1
	titleBarReachableHeight        = 40
	processQueryLimitedInformation = 0x1000
)

var (
	user32  = windows.NewLazySystemDLL("user32.dll")
	shell32 = windows.NewLazySystemDLL("shell32.dll")
	kernel  = windows.NewLazySystemDLL("kernel32.dll")
	dwmapi  = windows.NewLazySystemDLL("dwmapi.dll")

	procRegisterClassEx       = user32.NewProc("RegisterClassExW")
	procCreateWindowEx        = user32.NewProc("CreateWindowExW")
	procDefWindowProc         = user32.NewProc("DefWindowProcW")
	procDestroyWindow         = user32.NewProc("DestroyWindow")
	procGetMessage            = user32.NewProc("GetMessageW")
	procTranslateMessage      = user32.NewProc("TranslateMessage")
	procDispatchMessage       = user32.NewProc("DispatchMessageW")
	procPostMessage           = user32.NewProc("PostMessageW")
	procPostQuitMessage       = user32.NewProc("PostQuitMessage")
	procRegisterHotKey        = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey      = user32.NewProc("UnregisterHotKey")
	procLoadIcon              = user32.NewProc("LoadIconW")
	procExtractIconEx         = shell32.NewProc("ExtractIconExW")
	procShellNotifyIcon       = shell32.NewProc("Shell_NotifyIconW")
	procGetModuleHandle       = kernel.NewProc("GetModuleHandleW")
	procOpenProcess           = kernel.NewProc("OpenProcess")
	procCloseHandle           = kernel.NewProc("CloseHandle")
	procQueryFullProcessImage = kernel.NewProc("QueryFullProcessImageNameW")
	procDwmGetWindowAttribute = dwmapi.NewProc("DwmGetWindowAttribute")
	procGetCursorPos          = user32.NewProc("GetCursorPos")
	procSetForegroundWindow   = user32.NewProc("SetForegroundWindow")
	procCreatePopupMenu       = user32.NewProc("CreatePopupMenu")
	procAppendMenu            = user32.NewProc("AppendMenuW")
	procTrackPopupMenu        = user32.NewProc("TrackPopupMenu")
	procDestroyMenu           = user32.NewProc("DestroyMenu")
	procGetForegroundWindow   = user32.NewProc("GetForegroundWindow")
	procGetTopWindow          = user32.NewProc("GetTopWindow")
	procGetWindow             = user32.NewProc("GetWindow")
	procGetWindowThreadProcID = user32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible       = user32.NewProc("IsWindowVisible")
	procGetWindowText         = user32.NewProc("GetWindowTextW")
	procGetClassName          = user32.NewProc("GetClassNameW")
	procGetWindowRect         = user32.NewProc("GetWindowRect")
	procSetWindowPos          = user32.NewProc("SetWindowPos")
	procMonitorFromWindow     = user32.NewProc("MonitorFromWindow")
	procGetMonitorInfo        = user32.NewProc("GetMonitorInfoW")
	procIsIconic              = user32.NewProc("IsIconic")
	procIsZoomed              = user32.NewProc("IsZoomed")
	procShowWindow            = user32.NewProc("ShowWindow")

	trayMu       sync.RWMutex
	activeAgent  *WindowsAgent
	windowProcCB = syscall.NewCallback(windowProc)
)

type hotkeyReq struct {
	id     uint32
	value  string
	mods   uint32
	vk     uint32
	result chan error
}

type registeredHotkey struct {
	mods   uint32
	vk     uint32
	active bool
}

type quickPickCommand struct {
	presetID string
	target   uintptr
}

type WindowsAgent struct {
	app *App

	mu              sync.RWMutex
	config          Config
	presetByCmd     map[uint32]string
	quickPickByCmd  map[uint32]quickPickCommand
	hwnd            windows.Handle
	hIcon           windows.Handle
	resizeHotkey    registeredHotkey
	quickPickHotkey registeredHotkey
	stopped         bool
	hotkeyCh        chan hotkeyReq // dispatches RegisterHotKey to the message-loop OS thread
}

type point struct {
	X int32
	Y int32
}

type rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type resizeOutcome struct {
	requestedWidth  int32
	requestedHeight int32
	achievedWidth   int32
	achievedHeight  int32
}

type windowState struct {
	minimized bool
	maximized bool
}

func (s windowState) requiresRestore() bool {
	return s.minimized || s.maximized
}

func classifyResizeOutcome(preset Preset, achieved rect) resizeOutcome {
	return resizeOutcome{
		requestedWidth:  int32(preset.Width),
		requestedHeight: int32(preset.Height),
		achievedWidth:   achieved.Right - achieved.Left,
		achievedHeight:  achieved.Bottom - achieved.Top,
	}
}

func (o resizeOutcome) isExact() bool {
	return absInt32(o.requestedWidth-o.achievedWidth) <= resizeTolerance &&
		absInt32(o.requestedHeight-o.achievedHeight) <= resizeTolerance
}

func (o resizeOutcome) constrainedError(title string) error {
	return fmt.Errorf("%s constrained resize: requested %dx%d, achieved %dx%d",
		title, o.requestedWidth, o.requestedHeight, o.achievedWidth, o.achievedHeight)
}

func resizePosition(current rect, requestedWidth, requestedHeight int32, workArea rect, center bool) point {
	position := point{X: current.Left, Y: current.Top}
	if center {
		position.X = workArea.Left + ((workArea.Right-workArea.Left)-requestedWidth)/2
		position.Y = workArea.Top + ((workArea.Bottom-workArea.Top)-requestedHeight)/2
	}

	position.X = clampInt32(
		position.X,
		workArea.Left,
		maxInt32(workArea.Right-requestedWidth, workArea.Left),
	)
	position.Y = clampInt32(
		position.Y,
		workArea.Top,
		maxInt32(workArea.Bottom-titleBarReachableHeight, workArea.Top),
	)
	return position
}

func clampInt32(value, minimum, maximum int32) int32 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func maxInt32(left, right int32) int32 {
	if left > right {
		return left
	}
	return right
}

type message struct {
	Hwnd    windows.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
}

type wndClassEx struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     windows.Handle
	HIcon         windows.Handle
	HCursor       windows.Handle
	HbrBackground windows.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       windows.Handle
}

type notifyIconData struct {
	CbSize            uint32
	HWnd              windows.Handle
	UID               uint32
	UFlags            uint32
	UCallbackMessage  uint32
	HIcon             windows.Handle
	SzTip             [128]uint16
	DwState           uint32
	DwStateMask       uint32
	SzInfo            [256]uint16
	UTimeoutOrVersion uint32
	SzInfoTitle       [64]uint16
	DwInfoFlags       uint32
	GuidItem          windows.GUID
	HBalloonIcon      windows.Handle
}

type monitorInfo struct {
	CbSize    uint32
	RcMonitor rect
	RcWork    rect
	DwFlags   uint32
}

func NewPlatformAgent(app *App) PlatformAgent {
	app.mu.RLock()
	config := app.config.Clone()
	app.mu.RUnlock()
	return &WindowsAgent{
		app:            app,
		config:         config,
		presetByCmd:    map[uint32]string{},
		quickPickByCmd: map[uint32]quickPickCommand{},
		hotkeyCh:       make(chan hotkeyReq, 1),
	}
}

func (w *WindowsAgent) Start() error {
	started := make(chan error, 1)
	go w.run(started)
	if err := <-started; err != nil {
		return err
	}
	return w.ApplySettings(w.config)
}

func (w *WindowsAgent) Stop() {
	w.mu.Lock()
	if w.stopped {
		w.mu.Unlock()
		return
	}
	w.stopped = true
	hwnd := w.hwnd
	w.mu.Unlock()

	if hwnd != 0 {
		w.deleteTrayIcon()
		_, _, _ = procUnregisterHotKey.Call(uintptr(hwnd), hotkeyID)
		_, _, _ = procUnregisterHotKey.Call(uintptr(hwnd), quickPickHotkeyID)
		_, _, _ = procDestroyWindow.Call(uintptr(hwnd))
	}
}

func (w *WindowsAgent) ApplySettings(config Config) error {
	resizeMods, resizeVK, err := parseHotkey(config.Hotkey)
	if err != nil {
		return err
	}
	quickPickMods, quickPickVK, err := parseHotkey(config.QuickPickHotkey)
	if err != nil {
		return err
	}

	w.mu.Lock()
	hwnd := w.hwnd
	needsResizeHotkey := !w.resizeHotkey.active ||
		w.resizeHotkey.mods != resizeMods ||
		w.resizeHotkey.vk != resizeVK
	needsQuickPickHotkey := !w.quickPickHotkey.active ||
		w.quickPickHotkey.mods != quickPickMods ||
		w.quickPickHotkey.vk != quickPickVK
	w.mu.Unlock()

	// RegisterHotKey must run on the OS thread that owns the hidden window.
	// We send the request through hotkeyCh and wake the message loop via PostMessage.
	if hwnd != 0 {
		requests := make([]hotkeyReq, 0, 2)
		if needsResizeHotkey {
			requests = append(requests, hotkeyReq{
				id:     hotkeyID,
				value:  config.Hotkey,
				mods:   resizeMods,
				vk:     resizeVK,
				result: make(chan error, 1),
			})
		}
		if needsQuickPickHotkey {
			requests = append(requests, hotkeyReq{
				id:     quickPickHotkeyID,
				value:  config.QuickPickHotkey,
				mods:   quickPickMods,
				vk:     quickPickVK,
				result: make(chan error, 1),
			})
		}
		for _, req := range requests {
			w.hotkeyCh <- req
			_, _, _ = procPostMessage.Call(uintptr(hwnd), wmApplyHotkey, 0, 0)
			if err := <-req.result; err != nil {
				return err
			}
		}
	}

	// Only write autostart after the hotkey is confirmed working.
	if err := setAutoStart(config.AutoStart); err != nil {
		return err
	}

	w.mu.Lock()
	w.config = config.Clone()
	w.mu.Unlock()

	w.updateTrayIcon()
	return nil
}

func (w *WindowsAgent) run(started chan<- error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	trayMu.Lock()
	activeAgent = w
	trayMu.Unlock()
	defer func() {
		trayMu.Lock()
		if activeAgent == w {
			activeAgent = nil
		}
		trayMu.Unlock()
	}()

	hwnd, err := createHiddenWindow()
	if err != nil {
		started <- err
		return
	}

	// Extract the custom app icon embedded in this exe at the small tray size.
	// ExtractIconEx is more reliable than LoadIcon+resource-ID lookup since it
	// works regardless of which resource ID Wails assigns to the icon group.
	var hSmallIcon windows.Handle
	exePath, _ := os.Executable()
	lpszFile, _ := windows.UTF16PtrFromString(exePath)
	count, _, _ := procExtractIconEx.Call(
		uintptr(unsafe.Pointer(lpszFile)),
		0, // first icon group, index 0
		0, // skip large icon
		uintptr(unsafe.Pointer(&hSmallIcon)),
		1,
	)
	var hIcon windows.Handle
	if count > 0 && hSmallIcon != 0 {
		hIcon = hSmallIcon
	} else {
		// Fallback: generic Windows application icon
		icon, _, _ := procLoadIcon.Call(0, 32512)
		hIcon = windows.Handle(icon)
	}
	w.mu.Lock()
	w.hwnd = hwnd
	w.hIcon = hIcon
	w.mu.Unlock()

	if err := w.addTrayIcon(); err != nil {
		started <- err
		return
	}

	started <- nil

	var msg message
	for {
		ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) == -1 || ret == 0 {
			return
		}
		_, _, _ = procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		_, _, _ = procDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func createHiddenWindow() (windows.Handle, error) {
	className, err := windows.UTF16PtrFromString("ResizeMeTrayWindow")
	if err != nil {
		return 0, err
	}
	instance, _, _ := procGetModuleHandle.Call(0)
	wc := wndClassEx{
		CbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		LpfnWndProc:   windowProcCB,
		HInstance:     windows.Handle(instance),
		LpszClassName: className,
	}
	atom, _, registerErr := procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 && registerErr != windows.ERROR_CLASS_ALREADY_EXISTS {
		return 0, fmt.Errorf("register tray window: %w", registerErr)
	}

	hwnd, _, createErr := procCreateWindowEx.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(className)),
		0,
		0, 0, 0, 0,
		0,
		0,
		instance,
		0,
	)
	if hwnd == 0 {
		return 0, fmt.Errorf("create tray window: %w", createErr)
	}
	return windows.Handle(hwnd), nil
}

func windowProc(hwnd uintptr, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	trayMu.RLock()
	agent := activeAgent
	trayMu.RUnlock()

	if agent != nil {
		switch msg {
		case wmApplyHotkey:
			// Drain one pending hotkey request and register it on this OS thread.
			select {
			case req := <-agent.hotkeyCh:
				req.result <- agent.registerHotkey(req.id, req.value, req.mods, req.vk)
			default:
			}
			return 0
		case wmShowMenu:
			// TrackPopupMenu must run on the thread that owns the window.
			agent.showMenu()
			return 0
		case wmShowQuickPick:
			// TrackPopupMenu must run on the thread that owns the window.
			agent.showQuickPickMenu()
			return 0
		case wmTrayIcon:
			if lParam == wmLButtonUp || lParam == wmRButtonUp || lParam == wmContextMenu {
				_, _, _ = procPostMessage.Call(uintptr(hwnd), wmShowMenu, 0, 0)
				return 0
			}
		case wmHotkey:
			if wParam == hotkeyID {
				go func() {
					if err := agent.app.ResizeNow(); err != nil {
						agent.Notify("ResizeMe", err.Error(), true)
					}
				}()
				return 0
			}
			if wParam == quickPickHotkeyID {
				_, _, _ = procPostMessage.Call(hwnd, wmShowQuickPick, 0, 0)
				return 0
			}
		case wmCommand:
			go agent.handleCommand(uint32(wParam & 0xffff))
			return 0
		case wmDestroy:
			agent.deleteTrayIcon()
			_, _, _ = procUnregisterHotKey.Call(hwnd, hotkeyID)
			_, _, _ = procUnregisterHotKey.Call(hwnd, quickPickHotkeyID)
			_, _, _ = procPostQuitMessage.Call(0)
			return 0
		}
	}

	ret, _, _ := procDefWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

func (w *WindowsAgent) registerHotkey(id uint32, value string, mods uint32, vk uint32) error {
	w.mu.RLock()
	hwnd := w.hwnd
	previous := w.registeredHotkey(id)
	w.mu.RUnlock()
	if hwnd == 0 {
		return nil
	}

	// Temporarily unregister to free the ID so we can re-register.
	// We'll restore the old hotkey if the new one fails.
	if previous.active {
		_, _, _ = procUnregisterHotKey.Call(uintptr(hwnd), uintptr(id))
	}

	ret, _, err := procRegisterHotKey.Call(uintptr(hwnd), uintptr(id), uintptr(mods|modNoRepeat), uintptr(vk))
	if ret == 0 {
		// New hotkey failed — restore the previous one so the user isn't left
		// without a working hotkey.
		if previous.active {
			_, _, _ = procRegisterHotKey.Call(uintptr(hwnd), uintptr(id), uintptr(previous.mods|modNoRepeat), uintptr(previous.vk))
		}
		return fmt.Errorf("register %s hotkey %s: %w", hotkeyName(id), value, err)
	}

	w.mu.Lock()
	w.setRegisteredHotkey(id, registeredHotkey{mods: mods, vk: vk, active: true})
	w.mu.Unlock()
	return nil
}

func (w *WindowsAgent) registeredHotkey(id uint32) registeredHotkey {
	if id == quickPickHotkeyID {
		return w.quickPickHotkey
	}
	return w.resizeHotkey
}

func (w *WindowsAgent) setRegisteredHotkey(id uint32, hotkey registeredHotkey) {
	if id == quickPickHotkeyID {
		w.quickPickHotkey = hotkey
		return
	}
	w.resizeHotkey = hotkey
}

func hotkeyName(id uint32) string {
	if id == quickPickHotkeyID {
		return "quick-pick"
	}
	return "resize"
}

func parseHotkey(value string) (uint32, uint32, error) {
	parts := strings.Split(normalizeHotkeyText(value), "+")
	var mods uint32
	var key string
	for _, part := range parts {
		switch part {
		case "Ctrl":
			mods |= modControl
		case "Alt":
			mods |= modAlt
		case "Shift":
			mods |= modShift
		case "Win":
			mods |= modWin
		default:
			key = part
		}
	}
	if mods == 0 || key == "" {
		return 0, 0, fmt.Errorf("hotkey must include at least one modifier and one key")
	}
	if len(key) == 1 {
		ch := key[0]
		if (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
			return mods, uint32(ch), nil
		}
	}
	if strings.HasPrefix(key, "F") {
		number, err := strconv.Atoi(strings.TrimPrefix(key, "F"))
		if err == nil && number >= 1 && number <= 24 {
			return mods, uint32(0x70 + number - 1), nil
		}
	}
	return 0, 0, fmt.Errorf("unsupported hotkey key %q", key)
}

func (w *WindowsAgent) addTrayIcon() error {
	w.mu.RLock()
	hwnd := w.hwnd
	hIcon := w.hIcon
	w.mu.RUnlock()

	nid := notifyIconData{
		CbSize:           uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:             hwnd,
		UID:              1,
		UFlags:           nifMessage | nifIcon | nifTip,
		UCallbackMessage: wmTrayIcon,
		HIcon:            hIcon,
	}
	copyUTF16(nid.SzTip[:], w.tooltip())
	ret, _, err := procShellNotifyIcon.Call(nimAdd, uintptr(unsafe.Pointer(&nid)))
	if ret == 0 {
		return fmt.Errorf("add tray icon: %w", err)
	}
	return nil
}

func (w *WindowsAgent) updateTrayIcon() {
	w.mu.RLock()
	hwnd := w.hwnd
	hIcon := w.hIcon
	w.mu.RUnlock()
	if hwnd == 0 {
		return
	}

	nid := notifyIconData{
		CbSize:           uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:             hwnd,
		UID:              1,
		UFlags:           nifIcon | nifTip,
		UCallbackMessage: wmTrayIcon,
		HIcon:            hIcon,
	}
	copyUTF16(nid.SzTip[:], w.tooltip())
	_, _, _ = procShellNotifyIcon.Call(nimModify, uintptr(unsafe.Pointer(&nid)))
}

func (w *WindowsAgent) deleteTrayIcon() {
	w.mu.RLock()
	hwnd := w.hwnd
	w.mu.RUnlock()
	if hwnd == 0 {
		return
	}

	nid := notifyIconData{
		CbSize: uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:   hwnd,
		UID:    1,
	}
	_, _, _ = procShellNotifyIcon.Call(nimDelete, uintptr(unsafe.Pointer(&nid)))
}

func (w *WindowsAgent) tooltip() string {
	w.mu.RLock()
	config := w.config.Clone()
	w.mu.RUnlock()
	preset, ok := config.ActivePreset()
	if !ok {
		return "ResizeMe"
	}
	return fmt.Sprintf("ResizeMe - %s (%dx%d)", preset.Name, preset.Width, preset.Height)
}

func (w *WindowsAgent) Notify(title string, body string, warning bool) {
	w.mu.RLock()
	hwnd := w.hwnd
	hIcon := w.hIcon
	w.mu.RUnlock()
	if hwnd == 0 {
		return
	}

	flags := uint32(niifNone)
	if warning {
		flags = niifWarning
	}
	nid := notifyIconData{
		CbSize:      uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:        hwnd,
		UID:         1,
		UFlags:      nifInfo,
		HIcon:       hIcon,
		DwInfoFlags: flags,
	}
	copyUTF16(nid.SzInfoTitle[:], title)
	copyUTF16(nid.SzInfo[:], body)
	_, _, _ = procShellNotifyIcon.Call(nimModify, uintptr(unsafe.Pointer(&nid)))
}

type presetMenuItem struct {
	flags    uint32
	command  uint32
	label    string
	presetID string
}

func presetMenuItems(config Config, commandBase uint32) ([]presetMenuItem, uint32) {
	visiblePresets := config.VisiblePresets()
	items := make([]presetMenuItem, 0, len(visiblePresets)+4)
	nextPresetCommand := commandBase
	appendPreset := func(preset Preset) {
		flags := uint32(mfString)
		if preset.ID == config.ActivePresetID {
			flags |= mfChecked
		}
		items = append(items, presetMenuItem{
			flags:    flags,
			command:  nextPresetCommand,
			label:    fmt.Sprintf("%s  %dx%d", preset.Name, preset.Width, preset.Height),
			presetID: preset.ID,
		})
		nextPresetCommand++
	}

	favoriteSet := make(map[string]bool, len(config.FavoritePresetIDs))
	for _, id := range config.FavoritePresetIDs {
		favoriteSet[id] = true
	}

	hasFavorites := false
	for _, id := range config.FavoritePresetIDs {
		preset, ok := config.FindPreset(id)
		if !ok || config.IsPresetHidden(id) {
			continue
		}
		if !hasFavorites {
			items = append(items, presetMenuItem{flags: mfString | mfDisabled, label: "Favorites"})
			hasFavorites = true
		}
		appendPreset(preset)
	}

	otherPresets := make([]Preset, 0, len(visiblePresets))
	for _, preset := range visiblePresets {
		if favoriteSet[preset.ID] {
			continue
		}
		otherPresets = append(otherPresets, preset)
	}

	if hasFavorites && len(otherPresets) > 0 {
		items = append(items,
			presetMenuItem{flags: mfSeparator},
			presetMenuItem{flags: mfString | mfDisabled, label: "All Presets"},
		)
	} else if !hasFavorites {
		items = append(items, presetMenuItem{flags: mfString | mfDisabled, label: "Presets"})
	}
	for _, preset := range otherPresets {
		appendPreset(preset)
	}

	return items, nextPresetCommand
}

func (w *WindowsAgent) buildPresetMenu(menu uintptr, config Config, presetByCmd map[uint32]string, commandBase uint32) uint32 {
	appendMenu(menu, mfString|mfDisabled, 0, w.resizeTargetMenuLabel())
	appendMenu(menu, mfSeparator, 0, "")

	items, nextPresetCommand := presetMenuItems(config, commandBase)
	for _, item := range items {
		if item.presetID != "" {
			presetByCmd[item.command] = item.presetID
		}
		appendMenu(menu, item.flags, item.command, item.label)
	}
	return nextPresetCommand
}

func (w *WindowsAgent) showMenu() {
	w.mu.RLock()
	config := w.config.Clone()
	hwnd := w.hwnd
	w.mu.RUnlock()
	if hwnd == 0 {
		return
	}

	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer procDestroyMenu.Call(menu)

	presetByCmd := make(map[uint32]string, len(config.VisiblePresets()))
	w.buildPresetMenu(menu, config, presetByCmd, cmdPresetBase)

	w.mu.Lock()
	w.presetByCmd = presetByCmd
	w.mu.Unlock()

	appendMenu(menu, mfSeparator, 0, "")
	appendMenu(menu, mfString, cmdResize, fmt.Sprintf("Resize Now\t%s", config.Hotkey))
	appendMenu(menu, mfString|mfDisabled, 0, fmt.Sprintf("Resize hotkey: %s  |  Pick a size: %s", config.Hotkey, config.QuickPickHotkey))
	appendMenu(menu, mfSeparator, 0, "")

	centerFlags := uint32(mfString)
	if config.CenterAfterResize {
		centerFlags |= mfChecked
	}
	appendMenu(menu, centerFlags, cmdCenter, "Center after resize")
	appendMenu(menu, mfSeparator, 0, "")
	appendMenu(menu, mfString, cmdSettings, "Settings...")
	appendMenu(menu, mfString, cmdAbout, "About ResizeMe")
	appendMenu(menu, mfString, cmdQuit, "Quit ResizeMe")

	var pt point
	_, _, _ = procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	w.trackPopupMenu(menu, pt, hwnd)
}

func (w *WindowsAgent) showQuickPickMenu() {
	w.mu.RLock()
	config := w.config.Clone()
	hwnd := w.hwnd
	w.mu.RUnlock()
	if hwnd == 0 {
		return
	}

	target := w.resizeTarget()
	var targetRect rect
	targetAvailable := target != 0 && !w.isGuardedResizeTarget(target)
	if targetAvailable {
		if ret, _, _ := procGetWindowRect.Call(target, uintptr(unsafe.Pointer(&targetRect))); ret == 0 {
			targetAvailable = false
		}
	}

	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer procDestroyMenu.Call(menu)

	if !targetAvailable {
		appendMenu(menu, mfString|mfDisabled, 0, "No resizable window was in front")
		appendMenu(menu, mfString|mfDisabled, 0, "Click a window, then press the hotkey again")
		var cursor point
		_, _, _ = procGetCursorPos.Call(uintptr(unsafe.Pointer(&cursor)))
		w.mu.Lock()
		w.quickPickByCmd = map[uint32]quickPickCommand{}
		w.mu.Unlock()
		w.trackPopupMenu(menu, cursor, hwnd)
		return
	}

	presetByCmd := make(map[uint32]string, len(config.VisiblePresets()))
	w.buildPresetMenu(menu, config, presetByCmd, cmdQuickPickBase)
	quickPickByCmd := make(map[uint32]quickPickCommand, len(presetByCmd))
	for command, presetID := range presetByCmd {
		quickPickByCmd[command] = quickPickCommand{presetID: presetID, target: target}
	}

	appendMenu(menu, mfSeparator, 0, "")
	centerFlags := uint32(mfString)
	if config.CenterAfterResize {
		centerFlags |= mfChecked
	}
	appendMenu(menu, centerFlags, cmdCenter, "Center after resize")
	appendMenu(menu, mfString, cmdSettings, "Settings...")

	w.mu.Lock()
	w.quickPickByCmd = quickPickByCmd
	w.mu.Unlock()

	workArea, err := monitorWorkArea(windows.Handle(target))
	if err != nil {
		workArea = rect{}
	}
	w.trackPopupMenu(menu, quickPickMenuPoint(targetRect, workArea), hwnd)
}

func quickPickMenuPoint(target rect, workArea rect) point {
	point := point{X: target.Left, Y: target.Top}
	if workArea.Right <= workArea.Left || workArea.Bottom <= workArea.Top {
		return point
	}
	if point.X < workArea.Left {
		point.X = workArea.Left
	} else if point.X >= workArea.Right {
		point.X = workArea.Right - 1
	}
	if point.Y < workArea.Top {
		point.Y = workArea.Top
	} else if point.Y >= workArea.Bottom {
		point.Y = workArea.Bottom - 1
	}
	return point
}

func (w *WindowsAgent) trackPopupMenu(menu uintptr, pt point, hwnd windows.Handle) {
	_, _, _ = procSetForegroundWindow.Call(uintptr(hwnd))
	cmd, _, _ := procTrackPopupMenu.Call(
		menu,
		tpmRightButton|tpmReturnCmd|tpmNonotify,
		uintptr(int32ToUintptr(pt.X)),
		uintptr(int32ToUintptr(pt.Y)),
		0,
		uintptr(hwnd),
		0,
	)
	if cmd != 0 {
		// handleCommand calls Go/Wails methods — run off the message-loop thread.
		go w.handleCommand(uint32(cmd))
	}
}

func appendMenu(menu uintptr, flags uint32, command uint32, label string) {
	var labelPtr uintptr
	if label != "" {
		labelPtr = uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(label)))
	}
	_, _, _ = procAppendMenu.Call(menu, uintptr(flags), uintptr(command), labelPtr)
}

func (w *WindowsAgent) handleCommand(command uint32) {
	w.mu.RLock()
	config := w.config.Clone()
	presetID, isPreset := w.presetByCmd[command]
	quickPick, isQuickPick := w.quickPickByCmd[command]
	w.mu.RUnlock()

	switch {
	case isQuickPick:
		selected, err := w.app.SetActivePreset(quickPick.presetID)
		if err != nil {
			w.Notify("ResizeMe", err.Error(), true)
			return
		}
		preset, ok := selected.ActivePreset()
		if !ok {
			w.Notify("ResizeMe", fmt.Sprintf("unknown preset %q", quickPick.presetID), true)
			return
		}
		if err := w.resizeWindow(quickPick.target, preset, selected.CenterAfterResize); err != nil {
			w.Notify("ResizeMe", err.Error(), true)
		}
	case isPreset:
		if _, err := w.app.SetActivePreset(presetID); err != nil {
			w.Notify("ResizeMe", err.Error(), true)
		}
	case command == cmdCenter:
		if _, err := w.app.SetCenterAfterResize(!config.CenterAfterResize); err != nil {
			w.Notify("ResizeMe", err.Error(), true)
		}
	case command == cmdResize:
		if err := w.app.ResizeNow(); err != nil {
			w.Notify("ResizeMe", err.Error(), true)
		}
	case command == cmdSettings:
		w.app.ShowSettings()
	case command == cmdAbout:
		w.app.ShowAbout()
	case command == cmdQuit:
		w.app.Quit()
	}
}

func (w *WindowsAgent) ResizeActiveWindow(preset Preset, center bool) error {
	hwnd := w.resizeTarget()
	return w.resizeWindow(hwnd, preset, center)
}

func (w *WindowsAgent) resizeWindow(hwnd uintptr, preset Preset, center bool) error {
	if hwnd == 0 {
		return fmt.Errorf("no active window to resize")
	}

	w.mu.RLock()
	agentHwnd := w.hwnd
	w.mu.RUnlock()
	if windows.Handle(hwnd) == agentHwnd {
		return fmt.Errorf("ResizeMe cannot resize its own tray window")
	}

	var pid uint32
	_, _, _ = procGetWindowThreadProcID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid == uint32(os.Getpid()) {
		return fmt.Errorf("ResizeMe settings cannot be resized")
	}

	visible, _, _ := procIsWindowVisible.Call(hwnd)
	if visible == 0 {
		return fmt.Errorf("the active window is not resizable")
	}

	className := getWindowClass(windows.Handle(hwnd))
	switch className {
	case "Progman", "WorkerW", "Shell_TrayWnd", "Shell_SecondaryTrayWnd":
		return fmt.Errorf("the Windows desktop or taskbar cannot be resized")
	}

	state := windowState{}
	if ret, _, _ := procIsIconic.Call(hwnd); ret != 0 {
		state.minimized = true
	}
	if ret, _, _ := procIsZoomed.Call(hwnd); ret != 0 {
		state.maximized = true
	}
	if state.requiresRestore() {
		_, _, _ = procShowWindow.Call(hwnd, swRestore)
	}
	if state.minimized {
		if ret, _, _ := procIsIconic.Call(hwnd); ret != 0 {
			return fmt.Errorf("could not restore the minimized window")
		}
	}
	if state.maximized {
		if ret, _, _ := procIsZoomed.Call(hwnd); ret != 0 {
			return fmt.Errorf("could not restore the maximized window")
		}
	}

	var current rect
	if ret, _, err := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&current))); ret == 0 {
		return fmt.Errorf("read active window bounds: %w", err)
	}

	workArea, err := monitorWorkArea(windows.Handle(hwnd))
	if err != nil {
		return err
	}
	position := resizePosition(current, int32(preset.Width), int32(preset.Height), workArea, center)

	ret, _, err := procSetWindowPos.Call(
		hwnd,
		0,
		uintptr(int32ToUintptr(position.X)),
		uintptr(int32ToUintptr(position.Y)),
		uintptr(preset.Width),
		uintptr(preset.Height),
		swpNoZOrder|swpNoActivate,
	)
	if ret == 0 {
		title := getWindowTitle(windows.Handle(hwnd))
		if title == "" {
			title = "the active window"
		}
		return fmt.Errorf("could not resize %s: %w", title, err)
	}

	var achieved rect
	if ret, _, err := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&achieved))); ret == 0 {
		return fmt.Errorf("read resized window bounds: %w", err)
	}
	outcome := classifyResizeOutcome(preset, achieved)
	if !outcome.isExact() {
		title := getWindowTitle(windows.Handle(hwnd))
		if title == "" {
			title = "the active window"
		}
		return outcome.constrainedError(title)
	}
	return nil
}

func (w *WindowsAgent) resizeTargetMenuLabel() string {
	hwnd := w.resizeTarget()
	if hwnd == 0 {
		return "Target: No active window"
	}

	w.mu.RLock()
	agentHwnd := w.hwnd
	w.mu.RUnlock()

	var pid uint32
	_, _, _ = procGetWindowThreadProcID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	return targetMenuLabel(
		getWindowClass(windows.Handle(hwnd)),
		processNameForPID(pid),
		windows.Handle(hwnd) == agentHwnd || pid == uint32(os.Getpid()),
	)
}

func targetMenuLabel(className, processName string, isResizeMe bool) string {
	switch {
	case isResizeMe || className == "ResizeMeTrayWindow":
		return "Target: ResizeMe (not resizable)"
	case className == "Progman" || className == "WorkerW":
		return "Target: Windows desktop (not resizable)"
	case className == "Shell_TrayWnd" || className == "Shell_SecondaryTrayWnd":
		return "Target: Windows taskbar (not resizable)"
	case processName != "":
		return fmt.Sprintf("Target: %s", processName)
	default:
		return "Target: Active window"
	}
}

func (w *WindowsAgent) isGuardedResizeTarget(hwnd uintptr) bool {
	if hwnd == 0 {
		return true
	}

	w.mu.RLock()
	agentHwnd := w.hwnd
	w.mu.RUnlock()

	var pid uint32
	_, _, _ = procGetWindowThreadProcID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if windows.Handle(hwnd) == agentHwnd || pid == uint32(os.Getpid()) {
		return true
	}

	visible, _, _ := procIsWindowVisible.Call(hwnd)
	if visible == 0 {
		return true
	}
	minimized, _, _ := procIsIconic.Call(hwnd)
	if minimized != 0 {
		return true
	}

	switch getWindowClass(windows.Handle(hwnd)) {
	case "Progman", "WorkerW", "Shell_TrayWnd", "Shell_SecondaryTrayWnd", "ResizeMeTrayWindow":
		return true
	}
	return isWindowCloaked(windows.Handle(hwnd))
}

func (w *WindowsAgent) resizeTarget() uintptr {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return 0
	}

	w.mu.RLock()
	agentHwnd := w.hwnd
	w.mu.RUnlock()

	var foregroundPID uint32
	_, _, _ = procGetWindowThreadProcID.Call(hwnd, uintptr(unsafe.Pointer(&foregroundPID)))
	if windows.Handle(hwnd) != agentHwnd && foregroundPID != uint32(os.Getpid()) {
		return hwnd
	}

	hwnd, _, _ = procGetTopWindow.Call(0)
	for hwnd != 0 {
		var candidatePID uint32
		_, _, _ = procGetWindowThreadProcID.Call(hwnd, uintptr(unsafe.Pointer(&candidatePID)))
		visible, _, _ := procIsWindowVisible.Call(hwnd)
		minimized, _, _ := procIsIconic.Call(hwnd)
		className := getWindowClass(windows.Handle(hwnd))
		isSystemWindow := className == "Progman" || className == "WorkerW" || className == "Shell_TrayWnd" || className == "Shell_SecondaryTrayWnd"
		if windows.Handle(hwnd) != agentHwnd && candidatePID != uint32(os.Getpid()) && visible != 0 && minimized == 0 && !isSystemWindow && !isWindowCloaked(windows.Handle(hwnd)) {
			return hwnd
		}
		hwnd, _, _ = procGetWindow.Call(hwnd, gwHwndNext)
	}
	return 0
}

func isWindowCloaked(hwnd windows.Handle) bool {
	var cloaked uint32
	result, _, _ := procDwmGetWindowAttribute.Call(
		uintptr(hwnd),
		dwmwaCloaked,
		uintptr(unsafe.Pointer(&cloaked)),
		unsafe.Sizeof(cloaked),
	)
	return result == 0 && cloaked != 0
}

func monitorWorkArea(hwnd windows.Handle) (rect, error) {
	monitor, _, _ := procMonitorFromWindow.Call(uintptr(hwnd), monitorDefaultToNearest)
	if monitor == 0 {
		return rect{}, fmt.Errorf("could not find the active window's monitor")
	}
	info := monitorInfo{CbSize: uint32(unsafe.Sizeof(monitorInfo{}))}
	if ret, _, err := procGetMonitorInfo.Call(monitor, uintptr(unsafe.Pointer(&info))); ret == 0 {
		return rect{}, fmt.Errorf("read monitor bounds: %w", err)
	}
	return info.RcWork, nil
}

func getWindowTitle(hwnd windows.Handle) string {
	buffer := make([]uint16, 256)
	ret, _, _ := procGetWindowText.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	if ret == 0 {
		return ""
	}
	return windows.UTF16ToString(buffer[:ret])
}

func getWindowClass(hwnd windows.Handle) string {
	buffer := make([]uint16, 256)
	ret, _, _ := procGetClassName.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	if ret == 0 {
		return ""
	}
	return windows.UTF16ToString(buffer[:ret])
}

func processNameForPID(pid uint32) string {
	if pid == 0 {
		return ""
	}

	handle, _, _ := procOpenProcess.Call(processQueryLimitedInformation, 0, uintptr(pid))
	if handle == 0 {
		return ""
	}
	defer procCloseHandle.Call(handle)

	buffer := make([]uint16, 32768)
	length := uint32(len(buffer))
	ret, _, _ := procQueryFullProcessImage.Call(
		handle,
		0,
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(unsafe.Pointer(&length)),
	)
	if ret == 0 {
		return ""
	}
	return processNameFromPath(windows.UTF16ToString(buffer[:length]))
}

func processNameFromPath(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func setAutoStart(enabled bool) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open startup registration: %w", err)
	}
	defer key.Close()

	if !enabled {
		if err := key.DeleteValue("ResizeMe"); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove startup registration: %w", err)
		}
		return nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("read executable path: %w", err)
	}
	if err := key.SetStringValue("ResizeMe", quoteExecutablePath(exePath)); err != nil {
		return fmt.Errorf("write startup registration: %w", err)
	}
	return nil
}

func quoteExecutablePath(exePath string) string {
	return `"` + strings.Trim(exePath, `"`) + `"`
}

func copyUTF16(target []uint16, value string) {
	encoded := windows.StringToUTF16(value)
	if len(encoded) > len(target) {
		encoded = encoded[:len(target)]
		encoded[len(encoded)-1] = 0
	}
	copy(target, encoded)
}

func int32ToUintptr(value int32) uintptr {
	return uintptr(uint32(value))
}

func absInt32(value int32) int32 {
	if value < 0 {
		return -value
	}
	return value
}

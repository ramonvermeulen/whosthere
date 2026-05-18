package components

import (
	"fmt"
	"net"
	"slices"

	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/events"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

var _ UIComponent = &InterfacePicker{}

type InterfacePicker struct {
	*tview.List
	interfaces       []string
	currentInterface string
	initialized      bool
	emit             func(events.Event)
}

func NewInterfacePicker(emit func(events.Event)) *InterfacePicker {
	list := tview.NewList()
	list.ShowSecondaryText(false)

	ip := &InterfacePicker{
		List: list,
		emit: emit,
	}

	theme.RegisterPrimitive(list)
	return ip
}

func (ip *InterfacePicker) setupInputHandling() {
	ip.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Rune() == 'j' || event.Key() == tcell.KeyDown:
			nextIdx := ip.GetCurrentItem() + 1
			if nextIdx < len(ip.interfaces) {
				ip.SetCurrentItem(nextIdx)
			}
			return nil
		case event.Rune() == 'k' || event.Key() == tcell.KeyUp:
			prevIdx := ip.GetCurrentItem() - 1
			if prevIdx >= 0 {
				ip.SetCurrentItem(prevIdx)
			}
			return nil
		case event.Key() == tcell.KeyEnter:
			currentIdx := ip.GetCurrentItem()
			if currentIdx >= 0 && currentIdx < len(ip.interfaces) {
				selected := ip.interfaces[currentIdx]
				if selected != ip.currentInterface {
					ip.emit(events.InterfaceSelected{Name: selected})
				}
				ip.emit(events.HideView{})
			}
			return nil
		case event.Key() == tcell.KeyEsc || event.Rune() == 'q':
			ip.emit(events.HideView{})
			return nil
		}
		return event
	})
}

func (ip *InterfacePicker) Render(s state.ReadOnly) {
	newInterfaces := listUsableInterfaces()
	newCurrentInterface := s.CurrentInterface()

	interfacesChanged := !slices.Equal(ip.interfaces, newInterfaces)
	activeChanged := ip.currentInterface != newCurrentInterface

	if ip.initialized && !interfacesChanged && !activeChanged {
		return
	}

	savedIdx := ip.GetCurrentItem()

	ip.Clear()
	ip.interfaces = newInterfaces
	ip.currentInterface = newCurrentInterface

	ip.SetBorder(true).
		SetTitle(fmt.Sprintf(" Network Interfaces (%d) ", len(ip.interfaces))).
		SetTitleAlign(tview.AlignCenter).
		SetTitleColor(tview.Styles.TitleColor).
		SetBorderColor(tview.Styles.BorderColor).
		SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

	var activeIndex int
	for i, name := range ip.interfaces {
		displayName := name
		if desc := interfaceDescription(name); desc != "" {
			displayName += " (" + desc + ")"
		}
		if name == ip.currentInterface {
			displayName = "✓ " + displayName
			activeIndex = i
		}
		ip.AddItem(displayName, "", 0, nil)
	}

	if ip.initialized && !activeChanged && savedIdx >= 0 && savedIdx < len(ip.interfaces) {
		ip.SetCurrentItem(savedIdx)
	} else {
		ip.SetCurrentItem(activeIndex)
	}

	ip.initialized = true
	ip.setupInputHandling()
}

func listUsableInterfaces() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var result []string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil || len(addrs) == 0 {
			continue
		}
		hasIPv4 := false
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				hasIPv4 = true
				break
			}
		}
		if hasIPv4 {
			result = append(result, iface.Name)
		}
	}
	return result
}

func interfaceDescription(name string) string {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return ""
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			network := ipnet.IP.Mask(ipnet.Mask)
			ones, _ := ipnet.Mask.Size()
			return fmt.Sprintf("%s - %s/%d", ipnet.IP.String(), network.String(), ones)
		}
	}
	return ""
}





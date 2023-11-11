function init()
  m.listSettings        = m.top.findNode("listSettings")
  m.backgroundAutoplay  = m.top.findNode("backgroundAutoplay")
  m.radioAutoplay       = m.top.findNode("radioAutoplay")
  m.autoplayOn          = m.top.findNode("autoplayOn")
  m.autoplayOff         = m.top.findNode("autoplayOff")
  m.backgroundServer    = m.top.findNode("backgroundServer")
  m.labelServerAddress  = m.top.findNode("labelServerAddress")
  m.buttonServerAddress = m.top.findNode("buttonServerAddress")
  m.activeChild         = "listSettings"

  readSettings()
  m.listSettings.ObserveField("itemFocused", "OnSettingSelected")
  m.radioAutoplay.ObserveField("checkedItem", "OnAutoplaySelected")
  m.buttonServerAddress.ObserveField("buttonSelected", "OnServerAddressReset")
end function

function printSettings() as void
  autoplayString = "off"
  if m.global.autoPlay = true then autoplayString = "on"

  serverAddressString = m.global.serverAddress
  if serverAddressString = invalid then serverAddressString = "(invalid)"
  print "  Autoplay: " + autoplayString
  print "  Server  : " + serverAddressString
end function

function readSettings() as void
  if m.global.autoPlay = true then
    m.radioAutoplay.checkedItem = 0
  else
    m.radioAutoplay.checkedItem = 1
  end if

  serverAddressString = m.global.serverAddress
  if serverAddressString = invalid then serverAddressString = "(invalid)"
  m.labelServerAddress.text = "Server Address is: " + serverAddressString

  ' print "readSettings()"
  ' printSettings()
end function

function setFocusChild(child as string) as void
  m.activeChild = child
  if m.activeChild = "listSettings"        then m.listSettings.SetFocus(true)
  if m.activeChild = "radioAutoplay"       then m.radioAutoplay.SetFocus(true)
  if m.activeChild = "buttonServerAddress" then m.buttonServerAddress.SetFocus(true)
end function

function OnSettingSelected() as void
  settingsIndex = m.listSettings.ItemFocused
  settingsNode  = m.listSettings.content.GetChild(settingsIndex)

  if settingsNode.id = "headerAutoplay" then
    m.backgroundAutoplay.visible = true
    m.backgroundServer.visible   = false
  end if
  if settingsNode.id = "headerServerAddress" then
    m.backgroundAutoplay.visible = false
    m.backgroundServer.visible   = true
  end if
end function

function OnAutoplaySelected() as void
  if m.radioAutoplay.checkedItem = 0 then m.global.autoPlay = true
  if m.radioAutoplay.checkedItem = 1 then m.global.autoPlay = false
  printSettings()
end function

function OnServerAddressReset() as void
  m.top.serverAddressReset = true
end function

' Handle remote presses.
function OnKeyEvent(key as string, press as boolean) as boolean
  if press = false then return false

  if ((key = "right") or (key = "OK") or (key = "play")) and (m.activeChild = "listSettings") then
    settingsIndex = m.listSettings.ItemFocused
    settingsNode  = m.listSettings.content.GetChild(settingsIndex)
    if settingsNode.id = "headerAutoplay"      then setFocusChild("radioAutoplay")
    if settingsNode.id = "headerServerAddress" then setFocusChild("buttonServerAddress")
    return true
  end if

  if (key = "left") or (key = "back") then
    if m.activeChild = "listSettings" then
      m.top.appFocusMover = "listCategories"
    else
      setFocusChild("listSettings")
    end if
    return true
  end if

  return false
end function

function SetFocused(focused as boolean) as void
  if focused = false then return
  readSettings()
  setFocusChild("listSettings")
end function

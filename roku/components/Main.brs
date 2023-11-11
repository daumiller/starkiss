function init()
  ' Reference top level components
  m.effectStartup  = m.top.FindNode("effectStartup")
  m.videoPlayer    = m.top.FindNode("videoPlayer")
  m.listCategories = m.top.FindNode("listCategories")
  m.listMedia      = m.top.FindNode("listMedia")
  m.settings       = m.top.FindNode("settings")

  ' Set up global variables
  m.global.addFields({ autoPlay:true, serverAddress:"" })

  ' Play startup sound
  if m.effectStartup.loadStatus = "ready" then
    OnEffectStartup_Ready()
  else
    m.effectStartup.observeField("loadStatus", "OnEffectStartup_Ready")
  end if

  ' Set up event handlers
  m.listCategories.observeField("selectedCategory", "OnListCategories_Set")
  m.listCategories.observeField("appFocusMover",    "OnListCategories_FocusChanged")
  m.listCategories.observeField("errorMessage",     "OnListCategories_Error")
  m.listMedia.observeField("playerContent", "OnListMedia_Play")
  m.listMedia.observeField("appFocusMover", "OnListMedia_FocusChanged")
  m.listMedia.observeField("errorMessage",  "OnListMedia_Error")
  m.videoPlayer.observeField("state", "OnVideoPlayer_StateChanged")
  m.settings.observeField("appFocusMover", "OnSettings_FocusChanged")
  m.settings.observeField("errorMessage", "OnSettings_Error")
  m.settings.observeField("serverAddressReset", "OnSettings_ServerAddressReset")

  ' Set initial focus
  m.focusComponent = ""
  setFocusComponent("listCategories")

  ' Get server address before starting up network requests
  getServerAddress()
end function

function getServerAddress() as void
  ' attempt to read server address from registry
  ' if found, set global and continue on
  configuration = createObject("roRegistrySection", "configuration")
  if configuration.Exists("serverAddress") then
    m.global.setField("serverAddress", configuration.Read("serverAddress"))
    onServerAddress_Ready()
    return
  end if

  resetServerAddress(false)
end function

function resetServerAddress(failed as boolean) as void
  ' if not found, show dialog to enter server address
  dialogMessage = []
  if failed = true then dialogMessage.Push("Unable to connect to server at that address.")
  dialogMessage.Push("Enter the IP address of your Starkiss server:")

  serverDialog = createObject("roSGNode", "StandardKeyboardDialog")
  serverDialog.title   = "Starkiss Server"
  serverDialog.message = dialogMessage
  serverDialog.buttons = [ "OK" ]
  serverDialog.observeFieldScoped("buttonSelected", "onServerDialog_ButtonSelected")
  m.serverDialog = serverDialog
  m.top.dialog = serverDialog
end function

function onServerDialog_ButtonSelected() as void
  ' get address from dialog
  serverAddress = "http://" + m.serverDialog.text + ":4331"
  print "Address entered: " + serverAddress
  m.serverDialog.close = true
  m.serverDialog = invalid

  ' show connection dialog
  connectionDialog = createObject("roSGNode", "StandardMessageDialog")
  connectionDialog.title   = "Connecting..."
  connectionDialog.message = [ "Testing connection to server:", serverAddress ]
  connectionDialog.buttons = []
  m.connectionDialog = connectionDialog
  m.top.dialog = connectionDialog

  ' verify address is valid
  m.pingRequest = CreateObject("roSGNode", "PingRequest")
  m.pingRequest.SetFields({ address: serverAddress })
  m.pingRequest.ObserveField("success", "onServerAddress_Checked")
  m.pingRequest.control = "RUN"
end function

function onServerAddress_Checked() as void
  m.connectionDialog.close = true
  m.connectionDialog = invalid

  if m.pingRequest.success = false then
    resetServerAddress(true)
    return
  end if

  serverAddress = m.pingRequest.address
  m.global.setField("serverAddress", serverAddress)
  configuration = createObject("roRegistrySection", "configuration")
  configuration.Write("serverAddress", serverAddress)
  configuration.Flush()
  onServerAddress_Ready()
end function

function onServerAddress_Ready() as void
  ' Everything ready, start up network requests
  m.listMedia.visible = true
  m.settings.visible  = false
  setFocusComponent("listCategories")
  m.listCategories.callFunc("GetCategories")
end function

function setFocusComponent(component as string) as void
  m.focusComponent = component

  if m.focusComponent = "listCategories" then
    m.listCategories.visible = true
    m.videoPlayer.visible    = false
    m.listMedia.SetFields({ translation: "[320,0]" })
    m.settings.SetFields({ translation: "[320,0]" })
    m.listCategories.callFunc("SetExpanded", true)
    m.listCategories.callFunc("SetFocused", true)
    m.listMedia.callFunc("SetFocused", false)
    m.settings.callFunc("SetFocused", false)
    m.videoPlayer.SetFocus(false)
    return
  end if

  if m.focusComponent = "listMedia" then
    m.listCategories.visible = true
    m.listMedia.visible      = true
    m.settings.visible       = false
    m.videoPlayer.visible    = false
    m.listMedia.SetFields({ translation: "[64,0]" })
    m.listCategories.callFunc("SetExpanded", false)
    m.listCategories.callFunc("SetFocused", false)
    m.listMedia.callFunc("SetFocused", true)
    m.settings.callFunc("SetFocused", false)
    m.videoPlayer.SetFocus(false)
    return
  end if

  if m.focusComponent = "videoPlayer" then
    m.listCategories.visible = false
    m.listMedia.visible      = false
    m.settings.visible       = false
    m.videoPlayer.visible    = true
    m.listCategories.callFunc("SetExpanded", false)
    m.listCategories.callFunc("SetFocused", false)
    m.listMedia.callFunc("SetFocused", false)
    m.settings.callFunc("SetFocused", false)
    m.videoPlayer.SetFocus(true)
    return
  end if

  if m.focusComponent = "settings" then
    m.listCategories.visible = true
    m.listMedia.visible      = false
    m.settings.visible       = true
    m.videoPlayer.visible    = false
    m.settings.SetFields({ translation: "[64,0]" })
    m.listCategories.callFunc("SetExpanded", false)
    m.listCategories.callFunc("SetFocused", false)
    m.listMedia.callFunc("SetFocused", false)
    m.settings.callFunc("SetFocused", true)
    m.videoPlayer.SetFocus(false)
    return
  end if

  print "ERROR -> Main.brs: Unknown focus component: " + m.focusComponent
end function

function showMessageDialog(title as string, message as object) as void
  messageDialog = createObject("roSGNode", "StandardMessageDialog")
  messageDialog.title   = title
  messageDialog.message = message
  messageDialog.buttons = [ "OK" ]
  messageDialog.observeFieldScoped("buttonSelected", "onMessageDialog_ButtonSelected")
  m.top.dialog = messageDialog
end function
function onMessageDialog_ButtonSelected() as void
  m.top.dialog.close = true
end function

function OnListCategories_Set() as void
  m.listMedia.callFunc("SetParentId", m.listCategories.selectedCategory, true)
end function

function OnListCategories_FocusChanged() as void
  setFocusComponent(m.listCategories.appFocusMover)
end function

function OnListCategories_Error() as void
  if m.listCategories.errorMessage.count() < 1 then return
  showMessageDialog("Error", m.listCategories.errorMessage)
end function

function OnListMedia_FocusChanged() as void
  setFocusComponent(m.listMedia.appFocusMover)
end function

function OnListMedia_Play() as void
  print "Playing media: " + m.listMedia.playerContent.GetField("url")
  m.videoPlayer.content = m.listMedia.playerContent
  setFocusComponent("videoPlayer")
  m.videoPlayer.control = "play"
end function

function OnListMedia_Error() as void
  if m.listMedia.errorMessage.count() < 1 then return
  showMessageDialog("Error", m.listMedia.errorMessage)
end function

function OnSettings_FocusChanged() as void
  setFocusComponent(m.settings.appFocusMover)
end function

function OnSettings_Error() as void
  if m.settings.errorMessage.count() < 1 then return
  showMessageDialog("Error", m.settings.errorMessage)
end function

function OnSettings_ServerAddressReset() as void
  resetServerAddress(false)
end function

function OnVideoPlayer_StateChanged() as void
  if m.videoPlayer.state = "finished" then
    if m.global.autoPlay = true then
      if m.listMedia.callFunc("PlayNext") = true then return
    end if

    m.videoPlayer.control = "stop"
    setFocusComponent("listMedia")
  end if
end function

function OnEffectStartup_Ready() as void
  if m.effectStartup.loadStatus = "ready" then m.effectStartup.control = "play"
end function

function OnKeyEvent(key as String, press as Boolean) as Boolean
  if press = false then return false

  if m.focusComponent = "videoPlayer" then
    if key = "back" then
      m.videoPlayer.control = "stop"
      setFocusComponent("listMedia")
      return true
    end if
  end if

  return false
end function

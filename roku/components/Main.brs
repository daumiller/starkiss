function init()
  ' Reference top level components
  m.effectStartup  = m.top.FindNode("effectStartup")
  m.videoPlayer    = m.top.FindNode("videoPlayer")
  m.listCategories = m.top.FindNode("listCategories")
  m.listMedia      = m.top.FindNode("listMedia")

  ' Set initial focus
  m.focusComponent = ""
  setFocusComponent("listCategories")

  ' Set up event handlers
  m.listCategories.observeField("selectedCategory", "OnCategorySet")
  m.listCategories.observeField("appFocusMover", "OnAppFocusSet_listCategories")
  m.listMedia.observeField("playerContent", "OnPlayerContentSet")
  m.listMedia.observeField("appFocusMover", "OnAppFocusSet_listMedia")
  m.videoPlayer.observeField("state", "OnPlayerStateChanged")

  ' Play startup sound
  if m.effectStartup.loadStatus = "ready" then
    OnEffectStartupReady()
  else
    m.effectStartup.observeField("loadStatus", "OnEffectStartupReady")
  end if
end function

function setFocusComponent(component as string) as void
  m.focusComponent = component

  if m.focusComponent = "listCategories" then
    m.listCategories.visible = true
    m.listMedia.visible      = true
    m.videoPlayer.visible    = false
    m.listMedia.SetFields({ translation: "[320,0]" })
    m.listCategories.callFunc("SetExpanded", true)
    m.listCategories.callFunc("SetFocused", true)
    m.listMedia.callFunc("SetFocused", false)
    m.videoPlayer.SetFocus(false)
    return
  end if

  if m.focusComponent = "listMedia" then
    m.listCategories.visible = true
    m.listMedia.visible      = true
    m.videoPlayer.visible    = false
    m.listMedia.SetFields({ translation: "[64,0]" })
    m.listCategories.callFunc("SetExpanded", false)
    m.listCategories.callFunc("SetFocused", false)
    m.listMedia.callFunc("SetFocused", true)
    m.videoPlayer.SetFocus(false)
    return
  end if

  if m.focusComponent = "videoPlayer" then
    m.listCategories.visible = false
    m.listMedia.visible      = false
    m.videoPlayer.visible    = true
    m.listCategories.callFunc("SetExpanded", false)
    m.listCategories.callFunc("SetFocused", false)
    m.listMedia.callFunc("SetFocused", false)
    m.videoPlayer.SetFocus(true)
    return
  end if

  print "ERROR -> Main.brs: Unknown focus component: " + m.focusComponent
end function

function OnCategorySet() as void
  m.listMedia.callFunc("SetParentId", m.listCategories.selectedCategory, true)
end function

function OnAppFocusSet_listCategories() as void
  setFocusComponent(m.listCategories.appFocusMover)
end function

function OnAppFocusSet_listMedia() as void
  setFocusComponent(m.listMedia.appFocusMover)
end function

function OnPlayerContentSet() as void
  print "Playing media: " + m.listMedia.playerContent.GetField("url")
  m.videoPlayer.content = m.listMedia.playerContent
  setFocusComponent("videoPlayer")
  m.videoPlayer.control = "play"
end function

function OnPlayerStateChanged() as void
  if m.videoPlayer.state = "finished" then
    if m.global.autoPlay = true then
      if m.listMedia.callFunc("PlayNext") = true then return
    end if

    m.videoPlayer.control = "stop"
    setFocusComponent("listMedia")
  end if
end function

function OnEffectStartupReady() as void
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

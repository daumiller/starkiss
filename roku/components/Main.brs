function init()
  ' Reference top level components
  m.categories = m.top.FindNode("categories")
  m.media = m.top.FindNode("media")
  m.player = m.top.FindNode("player")
  m.sndstart = m.top.FindNode("sndstart")

  ' Set initial focus
  m.focusMode = ""
  SetFocusMode("categories")
  m.initialCategoryLoaded = false

  ' Set up event handlers
  m.categories.observeField("selectedCategory", "OnCategoryChanged")
  m.media.observeField("media", "OnMediaChanged")
  m.player.observeField("state", "OnPlayerStateChanged")

  ' Play startup sound
  if m.sndstart.loadStatus = "ready" then
    OnStartupSoundReady()
  else
    m.sndstart.observeField("loadStatus", "OnStartupSoundReady")
  end if
end function

function SetFocusMode(mode as String) as void
  if m.focusMode = mode then return
  m.focusMode = mode

  if m.focusMode = "categories" then
    m.categories.visible = true
    m.media.visible = true
    m.player.visible = false
    m.media.SetFields({ translation: "[320,0]" })
    m.categories.expanded = true
    m.media.expanded = false
    return
  end if

  if m.focusMode = "media" then
    m.categories.visible = true
    m.media.visible = true
    m.player.visible = false
    m.media.SetFields({ translation: "[64,0]" })
    m.categories.expanded = false
    m.media.expanded = true
    return
  end if

  if m.focusMode = "player" then
    m.categories.expanded = false
    m.media.expanded = false
    m.categories.visible = false
    m.media.visible = false
    m.player.visible = true
    return
  end if

  print "ERROR -> Main.brs: Unknown focus mode: " + mode
end function

function OnCategoryChanged() as void
  m.media.selectedCategory = m.categories.selectedCategory
  if m.initialCategoryLoaded = false then
    m.initialCategoryLoaded = true
  else
    SetFocusMode("media")
  end if
end function

function OnMediaChanged() as void
  print "Playing video: " + m.media.media.GetField("url")
  SetFocusMode("player")
  m.player.content = m.media.media
  m.player.control = "play"
  m.player.SetFocus(true)
end function

function OnPlayerStateChanged() as void
  if m.player.state = "finished" then
    ' TODO: if options.ContinuousPlayback = false then : SetFocusMode("media") : return : end if
    result = m.media.callFunc("PlayNextEntry")
    if result["next"] = false then SetFocusMode("media")
  end if
end function

function OnStartupSoundReady() as void
  if m.sndstart.loadStatus = "ready" then
    m.sndstart.control = "play"
  end if
end function

function OnKeyEvent(key as String, press as Boolean) as Boolean
  if press = false then return false

  if m.focusMode = "categories" then
    if key = "right" then
      SetFocusMode("media")
      return true
    end if
  end if

  if m.focusMode = "media" then
    if key = "left" then
      SetFocusMode("categories")
      return true
    end if
    if key = "back" then
      SetFocusMode("categories")
      return true
    end if
  end if

  if m.focusMode = "player" then
    if key = "back" then
      m.player.control = "stop"
      SetFocusMode("media")
      return true
    end if
  end if

  return false
end function

function init()
  ' Children
  m.listPosters = m.top.findNode("listPosters")
  m.listTitles  = m.top.findNode("listTitles")
  m.textStatus  = m.top.findNode("textStatus")

  ' Inernal Properites
  m.displayMode   = "status"
  m.listDepth     = 0
  m.listParents   = []
  m.listIndices   = []
  m.currentParent = ""
  m.playingIndex  = 0
  m.isFocused     = false
end function

' Set status message.
function setStatus(status as string) as void
  m.textStatus.text = status
end function

' Set listing display mode.
function setMode(mode as string) as void
  m.displayMode = mode

  if m.displayMode = "posters" then
    m.listPosters.visible = true
    m.listTitles.visible  = false
    m.textStatus.visible  = false
  end if
  if m.displayMode = "titles" then
    m.listPosters.visible = false
    m.listTitles.visible  = true
    m.textStatus.visible  = false
  end if
  if m.displayMode = "status" then
    m.listPosters.visible = false
    m.listTitles.visible  = false
    m.textStatus.visible  = true
  end if
end function

' Set parent ID and load listing. (If newRoot is true, reset list depth.)
function SetParentId(parentId as string, newRoot as boolean) as void
  setStatus("Loading...")
  setMode("status")
  if newRoot = true then
    m.listDepth   = 1
    m.listParents = [parentId]
    m.listIndices = [0]
  end if
  m.currentParent = parentId
  m.getRequest = CreateObject("roSGNode", "MediaListRequest")
  m.getRequest.SetFields({ id: m.currentParent })
  m.getRequest.ObserveField("content", "onListingLoaded")
  m.getRequest.control = "RUN"
end function

' Navigate down one level in listing.
function pushParent(parentId as string) as void
  if m.displayMode = "posters" then m.listIndices[m.listDepth - 1] = m.listPosters.itemFocused
  if m.displayMode = "titles"  then m.listIndices[m.listDepth - 1] = m.listTitles.itemFocused

  m.listDepth = m.listDepth + 1
  m.listParents.push(parentId)
  m.listIndices.push(0)

  SetParentId(parentId, false)
end function

' Navigate up one level in listing.
function popParent() as void
  if m.listDepth < 2 then return
  m.listDepth = m.listDepth - 1
  m.listParents.pop()
  m.listIndices.pop()

  SetParentId(m.listParents[m.listDepth - 1], false)
end function

' Set focus to current display element.
function SetFocused(focused as boolean) as void
  m.isFocused = focused
  if focused = false then return
  if m.displayMode = "posters" then m.listPosters.SetFocus(true)
  if m.displayMode = "titles"  then m.listTitles.SetFocus(true)
  if m.displayMode = "status"  then m.textStatus.SetFocus(true)
end function

' Listing network request completed. Populate list.
function onListingLoaded() as void
  if m.getRequest.error <> "" then
    print "Error loading listing: " + m.getRequest.error
    setStatus("Error loading listing: " + m.getRequest.error)
    setMode("status")
    return
  end if
  setStatus("")

  if (m.getRequest.listingType = "episodes") or (m.getRequest.listingType = "songs") then
    m.listTitles.content = m.getRequest.content
    setMode("titles")
    m.listTitles.jumpToItem = m.listIndices[m.listDepth - 1]
  else
    if m.getRequest.posterRatio = "2x3" then
      m.listPosters.SetFields({
        "basePosterSize"  : "[183, 275]",
        "numRows"         : "2",
        "numColumns"      : "6",
        "itemSpacing"     : "[16,32]",
        "translation"     : "[16,32]",
        "failedBitmapUri" : "pkg:/images/missing.small.2x3.png"
      })
    else ' m.getRequest.posterRatio = "1x1"
      m.listPosters.SetFields({
        "basePosterSize":"[200, 200]",
        "numRows":"3",
        "numColumns":"6",
        "failedBitmapUri":"pkg:/images/missing.small.1x1.png"
      })
    end if
    m.listPosters.content = m.getRequest.content
    setMode("posters")
    m.listPosters.jumpToItem = m.listIndices[m.listDepth - 1]
  end if

  if m.isFocused = true then SetFocused(true)
end function

' Play content.
function playContent(content as object) as void
  m.top.playerContent = content
end function

' Handle remote presses.
function OnKeyEvent(key as string, press as boolean) as boolean
  if press = false then return false

  if (key = "left") then
    m.top.appFocusMover = "listCategories"
    return true
  end if

  if (key = "back") then
    if m.listDepth > 1 then
      popParent()
    else
      m.top.appFocusMover = "listCategories"
    end if
    return true
  end if

  if (key = "down") then
    ' PosterGrid won't let you scroll down if the next column down is empty, even if next row isn't (completely) empty.
    ' Attempt to fix that:
    if m.displayMode = "posters" then
      if m.listPosters.itemFocused = (m.listPosters.content.GetChildCount() - 1) then return true
      m.listPosters.jumpToItem = (m.listPosters.content.GetChildCount() - 1)
      return true
    end if
    return false
  end if

  if (key = "OK") or (key = "play") then
    if m.displayMode = "status" then return true

    node_title = ""
    node_id    = ""
    node_type  = ""

    if m.displayMode = "posters" then
      index          = m.listPosters.itemFocused
      selectedNode   = m.listPosters.content.GetChild(index)
      node_title     = selectedNode.GetField("shortDescriptionLine1")
      node_id        = selectedNode.GetField("shortDescriptionLine2")
      node_type      = selectedNode.GetField("title")
      m.playingIndex = index
    else
      index          = m.listTitles.itemFocused
      selectedNode   = m.listTitles.content.GetChild(index)
      node_title     = selectedNode.GetField("title")
      node_id        = selectedNode.GetField("shortDescriptionLine2")
      node_type      = selectedNode.GetField("shortDescriptionLine1")
      m.playingIndex = index
    end if

    if (node_type <> "file-video") and (node_type <> "file-audio") then
      pushParent(node_id)
    else
      node_url      = m.global.serverAddress + "/media/" + node_id
      content       = CreateObject("roSGNode", "ContentNode")
      content.url   = node_url
      content.title = node_title
      playContent(content)
    end if

    return true
  end if

  return false
end function

' Play next entry in listing. (Used by auto-play.)
function PlayNext() as boolean
  if m.displayMode = "status" then return true

  node_title = ""
  node_id    = ""
  node_type  = ""

  if m.displayMode = "posters" then
    if m.playingIndex = (m.listPosters.content.GetChildCount() - 1) then return false
    m.playingIndex = m.playingIndex + 1
    m.listPosters.jumpToItem = m.playingIndex

    selectedNode   = m.listPosters.content.GetChild(m.playingIndex)
    node_title     = selectedNode.GetField("shortDescriptionLine1")
    node_id        = selectedNode.GetField("shortDescriptionLine2")
    node_type      = selectedNode.GetField("title")
  else
    if m.playingIndex = (m.listPosters.content.GetChildCount() - 1) then return false
    m.playingIndex = m.playingIndex + 1
    m.listPosters.jumpToItem = m.playingIndex

    selectedNode   = m.listTitles.content.GetChild(m.playingIndex)
    node_title     = selectedNode.GetField("title")
    node_id        = selectedNode.GetField("shortDescriptionLine2")
    node_type      = selectedNode.GetField("shortDescriptionLine1")
  end if

  if (node_type <> "file-video") and (node_type <> "file-audio") then return false

  node_url      = m.global.serverAddress + "/media/" + node_id
  content       = CreateObject("roSGNode", "ContentNode")
  content.url   = node_url
  content.title = node_title
  playContent(content)

  return true
end function

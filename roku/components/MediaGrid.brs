function init()
  m.grid = m.top.findNode("grid")
  m.list = m.top.findNode("list")
  m.mode = "grid"
  m.parentId = ""
  m.storedFocusIndex = 0
end function

function setMode(mode as string) as void
  m.mode = mode
end function

function setFocus() as void
  if m.mode = "grid" then
    m.grid.SetFocus(true)
  else
    m.list.SetFocus(true)
  end if
end function

function setContent(content as Object) as void
  if m.mode = "grid" then
    m.grid.content = content
  else
    m.list.content = content
  end if
end function

function setVisible(visible as Boolean) as void
  m.grid.visible = false
  m.list.visible = false
  if visible = true then
    if m.mode = "grid" then m.grid.visible = true
    if m.mode = "list" then m.list.visible = true
  end if
end function

function OnExpandedChanged() as void
  if m.top.expanded = true then setFocus()
end function

function OnCategoryChanged() as void
  m.getMediaRequest = CreateObject("roSGNode", "MediaListRequest")
  m.getMediaRequest.SetFields({ id: m.top.selectedCategory })
  m.getMediaRequest.ObserveField("content", "OnMediaLoaded")
  m.getMediaRequest.control = "RUN"
  setVisible(false)
end function

function OnMediaLoaded() as void
  setMode("grid")
  if m.getMediaRequest.listingType = "episodes" then setMode("list")
  if m.getMediaRequest.listingType = "songs"    then setMode("list")

  setContent(m.getMediaRequest.content)
  m.parentId = m.getMediaRequest.parentId
  if m.getMediaRequest.posterRatio = "2x3" then
    m.grid.SetFields({
      "basePosterSize"  : "[183, 275]",
      "numRows"         : "2",
      "numColumns"      : "6",
      "itemSpacing"     : "[16,32]",
      "translation"     : "[16,32]",
      "failedBitmapUri" : "pkg:/images/missing.small.2x3.png"
    })
  else ' m.getMediaRequest.posterRatio = "1x1"
    m.grid.SetFields({
      "basePosterSize":"[200, 200]",
      "numRows":"3",
      "numColumns":"6",
      "failedBitmapUri":"pkg:/images/missing.small.1x1.png"
    })
  end if
  setVisible(true)
  if m.top.expanded = true then setFocus()
end function

function OnKeyEvent(key as String, press as Boolean) as Boolean
  if press = false then return false

  if (key = "back") then
    if (m.parentId <> "") then
      if (m.parentId <> m.top.selectedCategory) then
        ' TODO: save & restore parent indices while navigating
        m.top.selectedCategory = m.parentId
      end if
      return true
    else
      ' fall-through, so Main.brs can use this to re-expand category selection
      return false
    end if
  end if

  if (key = "OK") or (key = "play") then
    node_title = ""
    node_id    = ""
    node_type  = ""

    if m.mode = "grid" then
      index = m.grid.itemFocused
      selectedNode = m.grid.content.GetChild(index)
      node_title = selectedNode.GetField("shortDescriptionLine1")
      node_id    = selectedNode.GetField("shortDescriptionLine2")
      node_type  = selectedNode.GetField("title")
      m.storedFocusIndex = index
    else
      index = m.list.itemFocused
      selectedNode = m.list.content.GetChild(index)
      node_title = selectedNode.GetField("title")
      node_id    = selectedNode.GetField("shortDescriptionLine2")
      node_type  = selectedNode.GetField("shortDescriptionLine1")
      m.storedFocusIndex = index
    end if

    ' print "Selected node type: " + node_type
    if (node_type <> "file-video") and (node_type <> "file-audio") then
      m.top.selectedCategory = node_id
      return true
    end if

    node_url = m.global.serverAddress + "/media/" + node_id

    content = CreateObject("roSGNode", "ContentNode")
    content.url = node_url
    content.title = node_title
    m.top.media = content

    return true
  end if

  return false
end function

function PlayNextEntry() as Object
  node_type  = ""
  node_id    = ""
  node_title = ""
  index = m.storedFocusIndex

  if m.mode = "grid" then
    if index = m.grid.content.GetChildCount() - 1 then
      result = { "next": false }
      return result
    end if
    m.storedFocusIndex = index + 1
    m.grid.jumpToItem = m.storedFocusIndex

    selectedNode = m.grid.content.GetChild(index + 1)
    node_title = selectedNode.GetField("shortDescriptionLine1")
    node_id    = selectedNode.GetField("shortDescriptionLine2")
    node_type  = selectedNode.GetField("title")
  end if

  if m.mode = "list" then
    if index = m.list.content.GetChildCount() - 1 then
      result = { "next": false }
      return result
    end if
    m.storedFocusIndex = index + 1
    m.list.jumpToItem = m.storedFocusIndex

    selectedNode = m.list.content.GetChild(index + 1)
    node_title = selectedNode.GetField("title")
    node_id    = selectedNode.GetField("shortDescriptionLine2")
    node_type  = selectedNode.GetField("shortDescriptionLine1")
  end if

  if (node_type <> "file-video") and (node_type <> "file-audio") then return { "next": false }

  node_url = m.global.serverAddress + "/media/" + node_id

  content = CreateObject("roSGNode", "ContentNode")
  content.url = node_url
  content.title = node_title
  m.top.media = content

  result = { "next": true }
  return result
end function

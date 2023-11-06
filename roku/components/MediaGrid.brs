function init()
  m.grid = m.top.findNode("grid")
  m.player = m.top.findNode("player")
  m.parentId = ""
end function

function OnExpandedChanged() as void
  if m.top.expanded = true then
    m.grid.SetFocus(true)
  end if
end function

function OnCategoryChanged() as void
  m.getMediaRequest = CreateObject("roSGNode", "MediaListRequest")
  m.getMediaRequest.SetFields({ id: m.top.selectedCategory })
  m.getMediaRequest.ObserveField("content", "OnMediaLoaded")
  m.getMediaRequest.control = "RUN"
  m.grid.visible = false
end function

function OnMediaLoaded() as void
  m.grid.content = m.getMediaRequest.content
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
  m.grid.visible = true
end function

function OnKeyEvent(key as String, press as Boolean) as Boolean
  if press = false then return false

  if (key = "back") then
    if (m.parentId <> "") then
      if (m.parentId <> m.top.selectedCategory) then
        m.top.selectedCategory = m.parentId
      end if
      return true
    else
      ' fall-through, so Main.brs can use this to re-expand category selection
      return false
    end if
  end if

  if (key = "OK") or (key = "play") then
    index = m.grid.itemFocused
    selectedNode = m.grid.content.GetChild(index)
    node_title = selectedNode.GetField("shortDescriptionLine1")
    node_id    = selectedNode.GetField("shortDescriptionLine2")
    node_type  = selectedNode.GetField("title")

    ' print "Selected node type: " + node_type
    if (node_type <> "file-video") and (node_type <> "file-audio") then
      m.top.selectedCategory = node_id
      return true
    end if

    node_url = m.global.serverAddress + "/media/" + node_id
    print "Playing " + node_title + " from " + node_url

    content = CreateObject("roSGNode", "ContentNode")
    content.url = node_url
    content.title = node_title
    m.top.media = content

    return true
  end if

  return false
end function

function PlayNextEntry() as void
  index = m.grid.itemFocused
  if index = m.grid.content.GetChildCount() - 1 then return
  m.grid.jumpToItem(index + 1)

  selectedNode = m.grid.content.GetChild(index + 1)
  node_title = selectedNode.GetField("shortDescriptionLine1")
  node_id    = selectedNode.GetField("shortDescriptionLine2")
  node_type  = selectedNode.GetField("title")

  if (node_type <> "file-video") and (node_type <> "file-audio") then return

  node_url = m.global.serverAddress + "/media/" + node_id
    print "Playing " + node_title + " from " + node_url

  content = CreateObject("roSGNode", "ContentNode")
  content.url = node_url
  content.title = node_title
  m.top.media = content
end function

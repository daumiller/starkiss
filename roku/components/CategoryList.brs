function init()
  m.top.selectedCategory = ""
  m.rect = m.top.FindNode("rect")
  m.icon = m.top.FindNode("icon")
  m.list = m.top.FindNode("list")

  m.catmap  = []
  m.getListingRequest = CreateObject("roSGNode", "CategoryListRequest")
  m.getListingRequest.ObserveField("listing", "OnCategoriesLoaded")
  m.getListingRequest.control = "RUN"
end function

function OnCategoriesLoaded() as void
  listContent = CreateObject("roSGNode", "ContentNode")
  listContent.setFields({ "role":"Content" })

  for each item in m.getListingRequest.listing
    m.catmap.push(item.id)
    itemContent = CreateObject("roSGNode", "ContentNode")
    itemContent.SetFields({ title: item["name"] })
    listContent.appendChild(itemContent)
  end for

  m.list.content = listContent

  ' during application startup, focus isn't set until we've loaded categories
  if m.top.expanded = true then m.list.SetFocus(true)
  ' auto-load first category
  if (m.top.selectedCategory = "") and (m.catmap.count() > 0) then
    m.top.selectedCategory = m.catmap[0]
  end if
end function

function OnExpandedChanged() as void
  selected_index = 0
  if m.top.selectedCategory <> "" then
    index_max = m.catmap.count() - 1
    for index = 0 to index_max step 1
      if m.catmap[index] = m.top.selectedCategory then
        selected_index = index
        exit for
      end if
    end for
  end if

  if m.top.expanded = true then
    m.rect.SetFields({ width: 320 })
    m.icon.visible = false
    m.list.visible = true
    m.list.jumpToItem = selected_index
    m.list.SetFocus(true)
  else
    m.rect.SetFields({ width: 64 })
    m.icon.visible = true
    m.list.visible = false
  end if
end function

function OnKeyEvent(key as String, press as Boolean) as Boolean
  if press = false then return false

  if (key = "OK") or (key = "play") then
    index = m.list.itemFocused
    m.top.selectedCategory = m.catmap[index]
    return true
  end if

  return false
end function

' sub FocusChanged()
'   if m.top.isInFocusChain() then
'     if m.rect.width = 320 then return
'     m.rect.SetFields({ width: 320 })
'     m.icon.visible = false
'     m.list.visible = true
'     m.list.setFocus(true)
'     m.list.jumpToItem = 0
'   else
'     if m.rect.width = 64 then return
'     m.rect.SetFields({ width: 64 })
'     m.icon.visible = true
'     m.list.visible = false
'   end if
' end sub

' current_content_count = list.content.getChildCount()
' if current_content_count = 1 then return
' itemContainer = CreateObject("roSGNode", "ContentNode")
' itemContainer.setFields({ "role":"Content" })
' item = CreateObject("roSGNode", "ContentNode")
' item.SetFields({ title: "X" })
' itemContainer.appendChild(item)
' list.content = itemContainer
' list.SetFields({ itemSize:[32,32], numRows:1 })

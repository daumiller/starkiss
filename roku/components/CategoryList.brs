function init()
  m.background     = m.top.FindNode("background")
  m.iconCollapsed  = m.top.FindNode("iconCollapsed")
  m.listCategories = m.top.FindNode("listCategories")

  m.categoryIds = []
  m.getListingRequest = CreateObject("roSGNode", "CategoryListRequest")
  m.getListingRequest.ObserveField("listing", "OnCategoriesLoaded")
  m.getListingRequest.control = "RUN"
end function

function OnCategoriesLoaded() as void
  listContent = CreateObject("roSGNode", "ContentNode")
  listContent.setFields({ "role":"Content" })

  for each item in m.getListingRequest.listing
    m.categoryIds.push(item.id)
    itemContent = CreateObject("roSGNode", "ContentNode")
    itemContent.SetFields({ title: item["name"] })
    listContent.appendChild(itemContent)
  end for

  ' Grab focus & set initial category
  m.listCategories.content = listContent
  m.listCategories.SetFocus(true)
  if m.categoryIds.count() > 0 then m.top.selectedCategory = m.categoryIds[0]
end function

function SetExpanded(expanded as boolean) as void
  if expanded = true then
    ' get index of currently selected category
    selected_index = 0
    index_max = m.categoryIds.count() - 1
    for index = 0 to index_max step 1
      if m.categoryIds[index] = m.top.selectedCategory then
        selected_index = index
        exit for
      end if
    end for

    m.background.SetFields({ width: 320 })
    m.iconCollapsed.visible = false
    m.listCategories.visible = true
    m.listCategories.jumpToItem = selected_index
  else
    m.background.SetFields({ width: 64 })
    m.iconCollapsed.visible = true
    m.listCategories.visible = false
  end if
end function

function SetFocused(focused as boolean) as void
  if focused = true then m.listCategories.SetFocus(true)
end function

function OnKeyEvent(key as String, press as Boolean) as Boolean
  if press = false then return false

  if (key = "right") then
    m.top.appFocusMover = "listMedia"
    return true
  end if

  if (key = "OK") or (key = "play") then
    index = m.listCategories.itemFocused
    m.top.selectedCategory = m.categoryIds[index]
    m.top.appFocusMover = "listMedia"
    return true
  end if

  return false
end function

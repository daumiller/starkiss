function init()
  m.background     = m.top.FindNode("background")
  m.iconCollapsed  = m.top.FindNode("iconCollapsed")
  m.listCategories = m.top.FindNode("listCategories")
  m.listControls   = m.top.FindNode("listControls")

  m.categoryIds = []
  m.focusedList = "categories"
end function

function GetCategories() as void
  m.getRequest = CreateObject("roSGNode", "CategoryListRequest")
  m.getRequest.ObserveField("listing", "OnCategoriesLoaded")
  m.getRequest.control = "RUN"
end function

function OnCategoriesLoaded() as void
  listContent = CreateObject("roSGNode", "ContentNode")
  listContent.setFields({ "role":"Content" })

  if m.getRequest.error.count() > 0 then
    m.listCategories.content = listContent
    for index = 0 to m.getRequest.error.count() - 1
      print "Error: " + m.getRequest.error[index]
    end for
    m.top.errorMessage = m.getRequest.error
    return
  end if

  for each item in m.getRequest.listing
    m.categoryIds.push(item.id)
    itemContent = CreateObject("roSGNode", "ContentNode")
    itemContent.SetFields({ title: item["name"] })
    listContent.appendChild(itemContent)
  end for

  ' Grab focus & set initial category
  m.listCategories.content = listContent
  focusList("categories")
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
    m.listControls.visible = true
  else
    m.background.SetFields({ width: 64 })
    m.iconCollapsed.visible = true
    m.listCategories.visible = false
    m.listControls.visible = false
  end if
end function

function SetFocused(focused as boolean) as void
  if focused = false then return
  focusList(m.focusedList)
end function

function focusList(list as string) as void
  m.focusedList = list
  if m.focusedList = "categories" then m.listCategories.SetFocus(true)
  if m.focusedList = "controls"   then m.listControls.SetFocus(true)
end function

function OnKeyEvent(key as String, press as Boolean) as Boolean
  if press = false then return false

  if key = "up" then
    if m.focusedList = "controls" then
      focusList("categories")
      m.listCategories.jumpToItem = m.listCategories.content.GetChildCount() - 1
      return true
    end if
  end if

  if key = "down" then
    if m.focusedList = "categories" then
      focusList("controls")
      m.listControls.jumpToItem = 0
      return true
    end if
  end if

  if (key = "right") or (key = "back") then
    if m.focusedList = "categories" then m.top.appFocusMover = "listMedia"
    if m.focusedList  = "controls" then
      index = m.listControls.itemFocused
      if index = 0 then m.top.appFocusMover = "settings"
    end if
    return true
  end if

  if (key = "OK") or (key = "play") then
    if m.focusedList = "categories" then
      index = m.listCategories.itemFocused
      m.top.selectedCategory = m.categoryIds[index]
      m.top.appFocusMover = "listMedia"
    end if
    if m.focusedList = "controls" then
      index = m.listControls.itemFocused
      if index = 0 then m.top.appFocusMover = "settings"
    end if
    return true
  end if

  return false
end function

function init()
  ' m.listPosters = invalid
end function

function OnKeyEvent(key as String, press as Boolean) as Boolean
  if press = false then return false
  m.top.keyPress = key
  return true
end function

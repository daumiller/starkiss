sub Main()
  screen = CreateObject("roSGScreen")
  m.port = CreateObject("roMessagePort")
  screen.setMessagePort(m.port)

  scene = screen.CreateScene("Main")
  globals = screen.getGlobalNode()
  globals.addFields({
    "serverAddress": "http://192.168.0.250:4331",
  })

  screen.show()
  scene.setFocus(true)

  print "entering loop"
  while(true)
      msg = wait(0, m.port)
      msgType = type(msg)
      print "Message type  : " ; msgType
      print "Message node  : " ; msg.GetNode()
      print "Message field : " ; msg.GetField()
      print "Message value : " ; msg.Int()
      if msgType = "roSGScreenEvent" then
        print "abc"
        if msg.isScreenClosed() then return
      end if
  end while
  print "exited loop"
end sub

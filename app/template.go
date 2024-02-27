package app

const configTemplate0 = ``

const configTemplate = `
var telegram = require("telegram"); 
var ps = new PersistentStorage("virtual-test-device-storage", {global: true});
{{ $slaves := .Slaves }}
{{ $channels := .Channels }}
{{- range $Slave := $slaves }}
{{- range $Channel := $channels }}
defineRule("wb-map12e_{{$Slave}}/Ch {{$Channel}} Total AP energy", {
  whenChanged: "wb-map12e_{{$Slave}}/Ch {{$Channel}} Total AP energy",
  then: function (newValue, devName, cellName) {
      if(ps['s{{$Slave}}ch{{$Channel}}'] !== 0 && !ps['s{{$Slave}}ch{{$Channel}}']){
          ps['s{{$Slave}}ch{{$Channel}}'] = newValue
          var msg = "значение устройства {} (канал {}) проинициализировано впервые - {}".format(devName, cellName, ps['s{{$Slave}}ch{{$Channel}}'])
          telegram.SendMessage(msg)
      }
      if (newValue >= ps['s{{$Slave}}ch{{$Channel}}']){
          ps['s{{$Slave}}ch{{$Channel}}'] = newValue
      } else {
          var msg = "значение устройства {} (канал {}) уменьшилось, было {}, стало {}".format(devName, cellName, ps['s{{$Slave}}ch{{$Channel}}'], newValue)
          telegram.SendMessage(msg)
      }
  }
});
trackMqtt("/devices/wb-map12e_{{$Slave}}/controls/Ch {{$Channel}} Total AP energy", function(message){
  log.info("trackMqtt INFO => {}".format(JSON.stringify(message)))
});
{{- end }}
{{- end }}

`

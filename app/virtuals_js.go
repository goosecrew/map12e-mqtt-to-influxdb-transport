package app

func virtuals() {
	type Content struct {
		Slaves   []string
		Channels []string
	}

	//используется для генерации wb-rules/virtual.js
	content := Content{
		Slaves:   slaves,
		Channels: channels,
	}
	tpl, err := template.New(`config`).Parse(configTemplate)
	if err != nil {
		log.Fatalln(err)
	}
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, content); err != nil {
		log.Fatalln(err)
	}
	fmt.Println(buf.String())

}

package core

import (
	"html/template"
	"io"
)

const formPostHTML = `<html lang="en">
<head>
    <title>Submit This Form</title>
</head>
<body onload="javascript:document.forms[0].submit()">
<form method="post" action="{{ .RedirectURL }}">
    {{ range $key,$value := .Parameters }}
    {{ range $parameter:= $value}}
    <input type="hidden" name="{{$key}}" value="{{$parameter}}"/>
    {{end}}
    {{ end }}
</form>
</body>
</html>
`

func (o *OAuth2) FormPostResponse(redirectTo string, rw io.Writer) {
	tp := template.Must(template.New("form_post").Parse(formPostHTML))
	_ = tp.Execute(rw, map[string]any{
		"RedirectURL": redirectTo,
		"Parameters":  map[string]any{},
	},
	)
}

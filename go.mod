module github.com/fractos/kritic

go 1.14

require (
	github.com/fatih/color v1.10.0
	github.com/inancgumus/screen v0.0.0-20190314163918-06e984b86ed3
	github.com/logrusorgru/aurora v2.0.3+incompatible
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v0.0.0-00010101000000-000000000000
)

replace k8s.io/client-go => k8s.io/client-go v0.19.2

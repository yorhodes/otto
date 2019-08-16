echo 'console.log(/a(b)(c(d)e)/.test("zabcdez"))' | go run otto/main.go
echo 'console.log("zabcdezzabcdez".match(/a(b)(c(d)e)/g))' | go run otto/main.go
echo 'console.log("bcdezzabcdez".search(/a(b)(c(d)e)/g))' | go run otto/main.go

package perm_test

const (
	testCA = `-----BEGIN CERTIFICATE-----
MIIDJDCCAgygAwIBAgIUHOVTOLLdOnP1LIVg+i+igGSFU84wDQYJKoZIhvcNAQEL
BQAwDTELMAkGA1UEAxMCY2EwHhcNMTgwNDA0MTY1MjAyWhcNMTkwNDA0MTY1MjAy
WjANMQswCQYDVQQDEwJjYTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEB
ANvLwpsN73NS7A76r36G+GWn9vBcWmep8/OLpIwX/p0G4O4aGVW7kf18Z9yRfe7Z
5gY8bys4cl+bJD9rq12tNcncPdXd32dhBiR1y+U0XEQOW2o8NrIsuK+EyaJnNSHp
Wj4VMt6ehX4CM+895p9m5Qtu/KZxj2zTEAJRvUAsOjcs6saEPIdAEbN5glpyJCZ2
NaLuYMr2xETt9AvMKgq6bLXxYRfL7QeypwzXcDnK4y0UDafaj2EJvCNNWGGSsTq7
nPLtJDaAx/eBkg73HNhC5Fl0Xy5Np9FIE1WoBy7J/Oa1xGRnCs8UOlnTzrUT/S5j
wEFO4Rd3vnKKniwIQ5UvP3UCAwEAAaN8MHowHQYDVR0OBBYEFPRXzreLNL/zdpCW
oAVxQ6q/QekhMEgGA1UdIwRBMD+AFPRXzreLNL/zdpCWoAVxQ6q/QekhoRGkDzAN
MQswCQYDVQQDEwJjYYIUHOVTOLLdOnP1LIVg+i+igGSFU84wDwYDVR0TAQH/BAUw
AwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAnH0nXh/ARLzR7dZrwc7Z9uQJHtBoq4nk
ncB1cGytJ3qD3QOki+qynOclEbqzdppNUGnxlB8rYbnPIZgJyZf69wDUwt8pRVpw
Lo+KU4UL/cfKZxyZvX8SgXqSN5VZfUvzjQfhwo8g5g3uhxUPipRyU3AbpezSksRx
kXa4eeLGVbgoGJs4YHwEpIi6f0Vb5LgSrIhF67+ySAMY2ko3i1f3iKvi7ONhUV9s
SWq6xndEYEPbdJVviU4x/uAQy6GEIt55YP4hx3L1C/NuVSuuPxe/PYmUykC+glsb
S1NZMlj9j/IFpBSFdHi1pNQDQSLdDrFXTpDO9XG95JWQNg+Ho6COLw==
-----END CERTIFICATE-----`

	testCert = `-----BEGIN CERTIFICATE-----
MIIDOzCCAiOgAwIBAgIUKrT9HVmj1pHaJMp9IVsouY/iibYwDQYJKoZIhvcNAQEL
BQAwDTELMAkGA1UEAxMCY2EwHhcNMTgwNDA0MjMxOTA3WhcNMTkwNDA0MjMxOTA3
WjAUMRIwEAYDVQQDEwlsb2NhbGhvc3QwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
ggEKAoIBAQCv+ToU8nEE3ZoR84dXsJLTag4AdYeXsBp5YBDHB9JhYUqAv/Q/7T7X
SAISZ/OUaqmZ/lU6K2X4YHSmLJjyS4DgAv9TjGTIlA6vk4HgR/6pjjKpG2AwSSNj
D2/oJ085ulMVBHoZyRuO7cafbJ+tXLO+X8XT6mvfL2sehtMvKMmywo1LcynMHdSJ
FWPajHax79sbyO1o7dq9JDJXJ/8j2kP4SKs+Y+8J4Hei0Iop6jPBB1GRGq6lKhkc
vbheb8BZXA4lPx6MaGvAKZDqp0vt23dnSwHe6lJusRlHn5fc1MHzJBUqjjvaSPcB
js3XcD3BCg1/6zlI1RWtN6fLQc0MClenAgMBAAGjgYswgYgwHQYDVR0OBBYEFERh
0BljrDOXaKxwkUODrySuhc/yMA8GA1UdEQQIMAaHBH8AAAEwSAYDVR0jBEEwP4AU
9FfOt4s0v/N2kJagBXFDqr9B6SGhEaQPMA0xCzAJBgNVBAMTAmNhghQc5VM4st06
c/UshWD6L6KAZIVTzjAMBgNVHRMBAf8EAjAAMA0GCSqGSIb3DQEBCwUAA4IBAQAL
xXk/DmsaR0LQUyjo35Pn9ASxd+46xdmBWBk/3TKgS5wqB3176evg5oqUfNAxC2Zk
PYC/su4VHLVjNLFu3Hm3yx27eARLGz/P3akPsAfAFZYY5dNEdXm/ecDgLC4YqHKL
L01Bum3hFj+VRkId7IjPzcXxkdwjUTELp84k75jP6f/jic82IRiGS1A1oX/1gUNf
NpCqJm4r/IzTp5TkeTyaXYzsEe0PbWM086W8kTnd86yJryDCgr3iWI1P9rEBvS2d
GipsRJWupwwZ3x4pReRGY7X69FnE1NkzgoJ/1Hay7QhJ2XqgOrLMAoqmKhpnb4Mz
zoLWTokJ7LC5c/iz3fSK
-----END CERTIFICATE-----`

	testCertKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEAr/k6FPJxBN2aEfOHV7CS02oOAHWHl7AaeWAQxwfSYWFKgL/0
P+0+10gCEmfzlGqpmf5VOitl+GB0piyY8kuA4AL/U4xkyJQOr5OB4Ef+qY4yqRtg
MEkjYw9v6CdPObpTFQR6Gckbju3Gn2yfrVyzvl/F0+pr3y9rHobTLyjJssKNS3Mp
zB3UiRVj2ox2se/bG8jtaO3avSQyVyf/I9pD+EirPmPvCeB3otCKKeozwQdRkRqu
pSoZHL24Xm/AWVwOJT8ejGhrwCmQ6qdL7dt3Z0sB3upSbrEZR5+X3NTB8yQVKo47
2kj3AY7N13A9wQoNf+s5SNUVrTeny0HNDApXpwIDAQABAoIBACXVRWF/hkDvi9JU
M0LtGYQOhxgeLJq2J9r0hmbkDZ0WF7h6jH65+Qq71oYyhzHfhLsw7Q3mirPNuQaJ
DBD7nqeKvW4u/vQIsOeapQ+dKuk1QzsMQa/f6ZXAmeSlnujgYEBiiAXHMP+Xq15X
MjVJevNxD80x1yNSxIr5nanD8SlXQu6+J6HNzUWhfjDzzmqJjA7A9kcYYorfPUv3
m1Pexdm1w/vtZyxPMtQFMuBhwnGi2ykfCKXb8cI480DdVZiJ52RYyNODAATD0I41
1BPjsS/o55Qkz8VkzCDVW8CwGQu92Q2LXj59EHLCf5PuiI7pBVv2G+CRXUl9VI07
Z3BpcvECgYEA6Yq3WOKkT7CWj7f9MIi+Pll8Q5AFAojtwagDx3apRf+dLDGDv9u3
O8fshiEE3ln6Dscb10sxmXRD4X6ZoXWmYIJeRKma7USTe0IGgrCN+UHd7Hwi1PYW
TqO/XdoLOzp56Uc+zBY8GJ2x1lnLzlLSJCj9BeBWBLzj8gUGZh/Q5I0CgYEAwOVO
XBYuNRguhnkUgAEmciEWGgVpTiuIKS+K/mGbW6Q6Tzw8K/OT5qgy7NvcmPcuKyIv
ENJY35+PbvzlX8qJYjsO1V/qwSq81USCqCQNVkEmc7x0NJLd1/JNX/Jdf6y6zLru
ax3/u9YE9/SS8qgTbGRr5XOFPwk1CkkxwqSM0gMCgYBXlJa3dZ8K530+/k+r4Mv/
U82fBKZsUe9fnWN1bNGEF5zYkuUGkR4BBDN2BXHu9K0q145gSambE5fVO4Xfn+A5
9wnlE/mumvX31kXcwtsrK4FPCyqA1Jx+9zdvubJWjtJjIj2xiXEWBiVH7jrY8AQw
XVKt3nhDpJaTD0FcEPhkjQKBgH6Kr3Intt5r47Gh4rnqhz8dx3MAk8mNM0DZiJRC
kfl3bi0mtc6bdy48r1PFFB1hIm93eGrPoy/oa98ClrLVmnTPi3ac+tMH52L9E72c
EQfBq6kHOzB2HISa1vmXdJDaTp0aEGhDAM5Ho3DKiFAZxMw5wLKAqyvkLWB3DebD
rgHjAoGAF4pzxwwUb94N5PQFndaLo2T9fMTWq65X30bKob1fUkAFisle6Wr/ElSA
ibWlQc+15aqPT66wSTk0VAIddzaC+wdTOUGFcseChAAb/PUQPOQXp6L4zv69ISSs
CLR1d7jdyijJXoSsLmQnE0vc1gbUxBfnIwr0UdCERmyc/Jy51GQ=
-----END RSA PRIVATE KEY-----`
)
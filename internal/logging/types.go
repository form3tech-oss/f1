package logging

type RegisterLogHookFunc func(scenario string)

var NoneRegisterLogHookFunc = func(string) {}

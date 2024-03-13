package logging

type RegisterLogHookFunc func(scenario string) error

func NoneRegisterLogHookFunc(string) error {
	return nil
}

package logutil

type StdLogAdapter struct {
	Logger Logger
}

func (adapter *StdLogAdapter) Write(p []byte) (n int, err error) {
	adapter.Logger.Warn(string(p))
	return len(p), nil
}

package waiter

type WaiterOption func(c *waiterCfg)

func CatchSignals() WaiterOption {
	return func(c *waiterCfg) {
		c.catchSignals = true
	}
}

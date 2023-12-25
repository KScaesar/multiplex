package multiplex

type MessageFunc[Message any] func(dto Message) error

type MessageDecorator[Message any] func(next MessageFunc[Message]) MessageFunc[Message]

type MessageChain[Message any] struct {
	chainAll []MessageDecorator[Message]
}

func (c *MessageChain[Message]) Link(fn MessageFunc[Message]) MessageFunc[Message] {
	return LinkFuncAndChain(fn, c.chainAll...)
}

func (c *MessageChain[Message]) AddChain(chains ...MessageDecorator[Message]) *MessageChain[Message] {
	c.chainAll = append(c.chainAll, chains...)
	return c
}

func (c *MessageChain[Message]) SetChain(chains []MessageDecorator[Message]) *MessageChain[Message] {
	c.chainAll = chains
	return c
}

func LinkFuncAndChain[Message any](handler MessageFunc[Message], chains ...MessageDecorator[Message]) MessageFunc[Message] {
	n := len(chains)
	for i := n - 1; 0 <= i; i-- {
		decorator := chains[i]
		handler = decorator(handler)
	}
	return handler
}

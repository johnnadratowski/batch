/*
The context library defines different call contexts that can be passed to controllers from the router
Mainly used for passing data to the controller from middleware
*/
package context

// Context struct.
type Context struct {
	IdentityID string
}

func (c *Context) GetIdentity() string {
	return c.IdentityID
}

func (c *Context) SetIdentity(identity string) {
	c.IdentityID = identity
}

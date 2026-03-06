package plc

import (
	"fmt"

	gologix "github.com/danomagnum/gologix"
)

type Client struct {
	path   string
	driver *gologix.Client
}

func NewClient(path string) *Client {
	return &Client{
		path: path,
	}
}

func (c *Client) Connect() error {
	c.driver = gologix.NewClient(c.path)

	if err := c.driver.Connect(); err != nil {
		return fmt.Errorf("не удалось подключиться: %w", err)
	}

	return nil
}

func (c *Client) Disconnect() {
	if c.driver != nil {
		c.driver.Disconnect()
	}
}

// Read читает тег в переданную переменную
func (c *Client) Read(tag string, result interface{}) error {
	if c.driver == nil {
		return fmt.Errorf("не подключены к PLC")
	}

	return c.driver.Read(tag, result)
}

func (c *Client) Write(tag string, value interface{}) error {
	if c.driver == nil {
		return fmt.Errorf("не подключены к PLC")
	}

	return c.driver.Write(tag, value)
}

func (c *Client) IsConnected() bool {
	if c.driver == nil {
		return false
	}
	return c.driver.Connected()
}

func (c *Client) GetPath() string {
	return c.path
}

func (c *Client) ReadMulti(data interface{}) error {
	return c.driver.ReadMulti(data)
}

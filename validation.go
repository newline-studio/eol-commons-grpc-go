package commons

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

type ScopedValidate interface {
	Struct(value any, key string, prepareValidation func(v *validator.Validate)) error
	StructPlain(value any) error
}

type validationContainer struct {
	list    map[string]*validator.Validate
	mu      sync.Mutex
	parents []*validator.Validate
}

func NewScopedValidate(parents ...*validator.Validate) ScopedValidate {
	return &validationContainer{
		list:    make(map[string]*validator.Validate),
		mu:      sync.Mutex{},
		parents: parents,
	}
}

func (c *validationContainer) resolveKey(key string, prepareValidation func(v *validator.Validate)) *validator.Validate {
	if v, ok := c.list[key]; ok {
		return v
	}
	v := validator.New()
	prepareValidation(v)

	c.mu.Lock()
	c.list[key] = v
	c.mu.Unlock()
	return v
}

func (c *validationContainer) Struct(value any, key string, prepareValidation func(v *validator.Validate)) error {
	for _, v := range append(c.parents, c.resolveKey(key, prepareValidation)) {
		if err := v.Struct(value); err != nil {
			return err
		}
	}
	return nil
}

func (c *validationContainer) StructPlain(value any) error {
	for _, v := range c.parents {
		if err := v.Struct(value); err != nil {
			return err
		}
	}
	return nil
}

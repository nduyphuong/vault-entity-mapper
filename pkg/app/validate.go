package app

import (
	"errors"

	"github.com/nduyphuong/vault-entity-mapper/pkg/config"
)

/*
*
validate currently only validate alias name

	Prior to creating any alias it is important to consider the cardinality of the alias' name, since there are potential security issues to be aware of.
	The main one revolves around alias reuse.
	It is possible for multiple authenticated entities to be bound to the same alias, and therefore gain access to all of its privileges.
	It is recommended, whenever possible, to create a unique alias for each entity.
	This is especially true in the case of machine generated entities.

*
*/
func validate(conf config.Config) error {
	m := map[string]bool{}
	for _, v := range conf.EntityAliases {
		if _, exist := m[v.Name]; exist {
			return errors.New("duplicated alias name")
		}
		m[v.Name] = true
	}
	return nil
}

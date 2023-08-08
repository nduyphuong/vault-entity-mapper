package app

import (
	"context"
	"fmt"

	client "github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"github.com/mitchellh/mapstructure"
	"github.com/nduyphuong/vault-entity-mapper/pkg/config"
	log "github.com/sirupsen/logrus"
)

func Run(conf config.Config) error {
	ctx := context.TODO()
	if err := validate(conf); err != nil {
		return err
	}
	c, err := client.New(
		client.WithEnvironment(),
	)
	if err != nil {
		return err
	}
	var eId []string
	for _, v := range conf.Entities {
		var d EntityLookUpResponse
		r, _ := c.Identity.EntityLookUp(ctx, schema.EntityLookUpRequest{
			Name: v.Name,
		})
		if r != nil {
			if r.Data != nil {
				config := &mapstructure.DecoderConfig{
					ErrorUnused: true,
					Result:      &d,
				}
				decoder, err := mapstructure.NewDecoder(config)
				if err != nil {
					log.Errorf("decode entity lookup response: %v", err)
				}
				decoder.Decode(r.Data)
			}
		}
		if v.Deactived {
			eId = append(eId, d.Id)
			continue
		}
		_, err := c.Identity.EntityCreate(ctx, schema.EntityCreateRequest{
			Disabled: v.Disabled,
			Metadata: v.Metadata,
			Id:       d.Id,
			Name:     v.Name,
			Policies: v.Policies,
		})
		if err != nil {
			log.Errorf("create entity %s: %v", v.Name, err)
		}
		log.Infof("entity %s was created", v.Name)
	}
	// delete if entity disabled in config file
	// safe delete in config file after disabled
	if len(eId) > 0 {
		for _, v := range eId {
			log.Warnf("entity %s disabled in config, deleting", v)
		}
		_, err := c.Identity.EntityBatchDelete(ctx, schema.EntityBatchDeleteRequest{
			EntityIds: eId,
		})
		if err != nil {
			log.Errorf("delete entity: %v", err)
		}
	}
	for _, v := range conf.EntityAliases {
		if v.Deactived {
			continue
		}
		_, err := c.Identity.AliasCreate(ctx, schema.AliasCreateRequest{
			Name:          v.Name,
			CanonicalId:   v.CanonicalId,
			MountAccessor: v.MountAccessor,
		})
		if err != nil {
			log.Errorf("create entity alias %s: %v", v.Name, err)
		}
		log.Infof("entity alias %s was created", v.Name)
	}
	eAtoI := map[string]map[string]interface{}{}
	r, _ := c.Identity.EntityListAliasesById(ctx)
	if r == nil {
		return nil
	}
	if r.Data != nil {
		for k, tmp := range r.Data["key_info"].(map[string]interface{}) {
			v := tmp.(map[string]interface{})["name"].(string)
			eAtoI[v] = map[string]interface{}{
				"Id":        k,
				"Deactived": false,
			}
		}
	}
	fmt.Printf("eAtoI: %v\n", eAtoI)
	for _, v := range conf.EntityAliases {
		if v.Deactived {
			if _, ok := eAtoI[v.Name]; ok {
				eAtoI[v.Name]["Deactived"] = true
			}
			continue
		}
		if _, ok := eAtoI[v.Name]; !ok {
			_, err := c.Identity.AliasCreate(ctx, schema.AliasCreateRequest{
				Name:          v.Name,
				CanonicalId:   v.CanonicalId,
				MountAccessor: v.MountAccessor,
				Id:            eAtoI[v.Name]["Id"].(string),
			})
			if err != nil {
				log.Errorf("update entity alias %s: %v", v.Name, err)
			}
		}
	}
	fmt.Printf("eAtoI: %v\n", eAtoI)
	for k, v := range eAtoI {
		if v["Deactived"] == true {
			if _, err := c.Identity.AliasReadById(ctx, v["Id"].(string)); err == nil {
				log.Warnf("entity aliases %s  disabled in config, deleting", k)
				if _, err := c.Identity.EntityDeleteAliasById(ctx, v["Id"].(string)); err != nil {
					log.Errorf("delete %s", k)
				}
			}
		}
	}
	return nil
}

package app

import (
	"context"

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
		if v.Deactived {
			r, _ := c.Identity.EntityLookUp(ctx, schema.EntityLookUpRequest{
				Name: v.Name,
			})
			if r != nil {
				if r.Data != nil {
					var d EntityLookUpResponse
					config := &mapstructure.DecoderConfig{
						ErrorUnused: true,
						Result:      &d,
					}
					decoder, err := mapstructure.NewDecoder(config)
					if err != nil {
						log.Errorf("decode entity lookup response: %v", err)
					}
					decoder.Decode(r.Data)
					eId = append(eId, d.Id)
				}
			}
			continue
		}
		_, err := c.Identity.EntityLookUp(ctx, schema.EntityLookUpRequest{
			Name: v.Name,
		})
		if err != nil {
			_, err := c.Identity.EntityCreate(ctx, schema.EntityCreateRequest{
				Disabled: v.Disabled,
				Id:       v.Id,
				Metadata: v.Metadata,
				Name:     v.Name,
				Policies: v.Policies,
			})
			if err != nil {
				log.Errorf("create entity %s: %v", v.Name, err)
			}
			log.Infof("created entity %s", v.Name)
		}
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
	toDel := map[string]string{}
	for _, v := range conf.EntityAliases {
		if v.Deactived {
			toDel[v.Name] = ""
			continue
		}
		_, err := c.Identity.AliasCreate(ctx, schema.AliasCreateRequest{
			Name:          v.Name,
			CanonicalId:   v.CanonicalId,
			MountAccessor: v.MountAccessor,
			Id:            v.Id,
		})
		if err != nil {
			log.Errorf("create entity alias %s: %v", v.Name, err)
		}
		log.Infof("created entity alias %s", v.Name)
	}
	r, err := c.Identity.EntityListAliasesById(ctx)
	if err != nil {
		log.Errorf("list entity aliases: %v", err)
		return err
	}

	for k, tmp := range r.Data["key_info"].(map[string]interface{}) {
		v := tmp.(map[string]interface{})["name"].(string)
		if _, ok := toDel[v]; ok {
			toDel[v] = k
		}
	}
	for k, v := range toDel {
		log.Warnf("entity aliases %s id: %s disabled in config, deleting", k, v)
		if _, err := c.Identity.EntityDeleteAliasById(ctx, v); err != nil {
			log.Errorf("delete %s id: %s", k, v)
		}
	}
	return nil
}

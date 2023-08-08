package config

type Config struct {
	Entities      []Entity      `mapstructure:"entities"`
	EntityAliases []EntityAlias `mapstructure:"entitiesAliases"`
}

type Entity struct {
	Name      string                 `mapstructure:"name"`
	Id        string                 `mapstructure:"id"`
	Metadata  map[string]interface{} `mapstructure:"metadata"`
	Policies  []string               `mapstructure:"policies"`
	Disabled  bool                   `mapstructure:"disabled"`
	Deactived bool                   `mapstructure:"deactived"`
}

type EntityAlias struct {
	MountAccessor string `mapstructure:"mountAccessor"`
	Name          string `mapstructure:"name"`
	CanonicalId   string `mapstructure:"canonicalId"`
	Id            string `mapstructure:"id"`
	Deactived     bool   `mapstructure:"deactived"`
}

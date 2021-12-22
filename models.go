package monk

import (
	"time"

	"github.com/outerjoin/do"
)

/*
type PreInsertHook interface {
	BeforeInsert(input Map) error
}

type PreUpdateHook interface {
	BeforeUpdate(input Map) error
}
*/

type MongoStore struct {
}

type MysqlStore struct {
}

var globalHook = struct {
	Timed
}{
	Timed{},
}

type OptionalBehaviors struct {
	Address    bool
	Coordinate bool
	Seo        bool
	File       bool
	Files      bool
	State      bool
	Dynamic    bool
}

type Address struct {
	// Name?
	// Phone?
	Street     string
	Street2    string
	Landmark   string
	PostalCode string
	CityID     string
	City       string
	StateID    string
	State      string
	CountryID  string
	Country    string
}

type Coordinate struct {
	Latitutde    float32
	Longitude    float32
	LocationName *string
}

type BusinessProcess struct {
	ProcessState string
}

func (BusinessProcess) BeforeInsert(input do.Map) error {
	return nil
}

func (BusinessProcess) BeforeUpdate(input do.Map) error {
	return nil
}

type Tagged struct {
	Tags []string
}

type Reference struct {
	Context string
	RefID   int
	RefUID  string
}

type Versioned struct {
	Version int
}

func (Versioned) BeforeInsert(input do.Map) error {
	return nil
}

func (Versioned) BeforeUpdate(input do.Map) error {
	return nil
}

type Active0 struct {
	Active bool `bson:"active" json:"active" index:"true" default:"0"`
}

type Active1 struct {
	Active bool `bson:"active" json:"active" index:"true" default:"1"`
}

type Timed struct {
	CreatedAt time.Time `bson:"created_at" json:"created_at" index:"true"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at" index:"true"`
}

func (Timed) BeforeInsert(input do.Map) error {
	return nil
}

func (Timed) BeforeUpdate(input do.Map) error {
	return nil
}

type Who struct {
	Username  *string
	UserID    *int
	UserUID   *string
	SessionID *string
}

type File struct {
	Source string
	Mime   string
	Size   uint
	Width  *int
	Height *int
}

type Files []File

type Seo struct {
	URL        *string
	URLHistory *[]string

	// Should title, keyword and description also
	// be a part of Metas?
	MetaTitle       *string
	MetaKeywords    *string
	MetaDescription *string

	Metas *map[string]string
}

type Activated1 struct {
	// All queries assume Active=1
	Active uint
}

func (Activated1) BeforeInsert(input do.Map) error {
	return nil
}

func (Activated1) BeforeUpdate(input do.Map) error {
	return nil
}

type Activated0 struct {
	// All queries assume Active=1
	Active uint
}

type Deletable struct {
	// All queries assume Deleted=0
	Deleted uint
}

func (Deletable) BeforeInsert(input do.Map) error {
	return nil
}

func (Deletable) BeforeUpdate(input do.Map) error {
	return nil
}

func (Deletable) BeforeDelete(input do.Map) error {
	return nil
}

type CustomFields struct {
	Custom *map[string]interface{}
}

type AttributeFields struct {
	Attributes *map[string]interface{}
}

type Api struct {
	ID      uint
	Type    string
	Version uint // -> Indicates which "type" is used (supports evolution of Blog types)
}

var API_BLOG_01 = Api{1, "blog", 1}

type TelemetryConfig struct {
	Provider string
	NewRelic *NewRelicConfig  `bson:"new_relic" json:"new_relic"`
	DataDog  *DataDogNewRelic `bson:"data_dog" json:"data_dog"`
}

type NewRelicConfig struct {
}

type DataDogNewRelic struct {
}

type User struct {
	UUID     string `bson:"_id" json:"uuid"`
	Username string `bson:"username" json:"username"`
	Password string `bson:"password" json:"password"`

	// Role (enum)

	AccountUUID string `bson:"account_uuid" json:"account_uuid"`

	Activated0
	CustomFields
	Tagged
	Timed
}

type Account struct {
	UUID string `bson:"_id" json:"uuid"`

	// AccountLevel (enum)

	Deletable
	CustomFields
	Tagged
	Timed
}

type Environment struct {
	UUID string `bson:"_id" json:"uuid"`
	Name string `bson:"name" json:"name"`

	AccountUUID string `bson:"account_uuid" json:"account_uuid"`

	TelemetryConfig *TelemetryConfig `bson:"telemetry_config" json:"telemetry_config"`

	Activated1
	Deletable
	CustomFields
	Tagged
	Timed
}

type Instance struct {
	UUID string `bson:"_id" json:"uuid"`
	Name string `bson:"name" json:"name"`

	EnvironmentUUID string `bson:"environment_uuid" json:"environment_uuid"`
	AccountUUID     string `bson:"account_uuid" json:"account_uuid"`

	Api         uint               `bson:"api" json:"api"`
	Options     *OptionalBehaviors `bson:"options" json:"options"`
	LogChanges  int                `bson:"log_changes" json:"log_changes"`
	AllowUpload int                `bson:"allow_upload" json:"allow_upload"`
	DoTelemetry int                `bson:"do_telemetry" json:"do_telemetry"`

	Activated1
	Deletable
	CustomFields
	Tagged
	Timed
}

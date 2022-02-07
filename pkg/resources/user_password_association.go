package resources

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Compact(d []string) []string {
	r := make([]string, 0)

	for _, v := range d {
		if v != "" {
			r = append(r, v)
		}
	}

	return r
}

func getId(d *schema.ResourceData) string {
	ids := []string{
		d.Get("user").(string),
		d.Get("secret_id").(string),
	}

	return strings.Join(Compact(ids), "-")
}

func getPassword(secretId string, session *session.Session) (string, error) {
	secretsManagerClient := secretsmanager.New(session)

	gsvi := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretId),
	}

	gsvo, err := secretsManagerClient.GetSecretValue(gsvi)

	if err != nil {
		return "", err
	}

	return *gsvo.SecretString, nil
}

func SnowflakeUserPasswordAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: SnowflakeUserPasswordAssociationCreate,
		ReadContext:   SnowflakeUserPasswordAssociationRead,
		UpdateContext: SnowflakeUserPasswordAssociationRead,
		DeleteContext: SnowflakeUserPasswordAssociationDelete,
		Schema: map[string]*schema.Schema{
			"user": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "name of user updating the password",
			},
			"secret_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "id of secret",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(60 * time.Second),
		},
	}
}

func SnowflakeUserPasswordAssociationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	username := d.Get("user").(string)
	secretId := d.Get("secret_id").(string)
	password, err := getPassword(secretId, getSession())

	if err != nil {
		return diag.FromErr(err)
	}

	client := m.(*sql.DB)

	tx, txBeginErr := client.Begin()
	if txBeginErr != nil {
		log.Println("error | SnowflakeUserPasswordAssociationCreate | txBeginErr |", txBeginErr)
		return diag.FromErr(txBeginErr)
	}

	alterPasswordStatement := fmt.Sprintf("ALTER USER %s SET PASSWORD = '%s'", username, password)
	if _, alterPasswordrErr := tx.Exec(alterPasswordStatement); alterPasswordrErr != nil {
		log.Println("error | SnowflakeUserPasswordAssociationCreate | alterPasswordrErr |", alterPasswordrErr)
		tx.Rollback()
		return diag.FromErr(alterPasswordrErr)
	}

	if txCommitErr := tx.Commit(); txCommitErr != nil {
		log.Println("error | SnowflakeUserPasswordAssociationCreate | txCommitErr |", txCommitErr)
		return diag.FromErr(txCommitErr)
	}

	d.SetId(getId(d))

	return diags
}

func SnowflakeUserPasswordAssociationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(getId(d))

	return diags
}

func SnowflakeUserPasswordAssociationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

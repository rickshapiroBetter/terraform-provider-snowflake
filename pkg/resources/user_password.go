package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func getSession() *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)

	if err == nil {
		return sess
	} else {
		fmt.Println(err.Error())
		return nil
	}
}

func generateRandomPassword(svc *secretsmanager.SecretsManager) (string, error) {
	gpi := &secretsmanager.GetRandomPasswordInput{
		ExcludePunctuation: aws.Bool(true),
		PasswordLength:     aws.Int64(32),
	}

	gpo, err := svc.GetRandomPassword(gpi)

	if err != nil {
		return "", err
	}

	return *gpo.RandomPassword, nil
}

func SnowflakeUserPassword() *schema.Resource {
	return &schema.Resource{
		CreateContext: SnowflakeUserPasswordCreate,
		ReadContext:   SnowflakeUserPasswordRead,
		UpdateContext: SnowflakeUserPasswordRead,
		DeleteContext: SnowflakeUserPasswordDelete,
		Schema: map[string]*schema.Schema{
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

func SnowflakeUserPasswordCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	secretsManager := secretsmanager.New(getSession())
	userPassword, err := generateRandomPassword(secretsManager)

	if err != nil {
		return diag.FromErr(err)
	}

	secretId := d.Get("secret_id").(string)

	psvi := &secretsmanager.PutSecretValueInput{
		SecretId:     aws.String(secretId),
		SecretString: aws.String(string(userPassword)),
	}

	_, err = secretsManager.PutSecretValue(psvi)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(secretId)

	return diags
}

func SnowflakeUserPasswordRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(d.Get("secret_id").(string))

	return diags
}

func SnowflakeUserPasswordDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

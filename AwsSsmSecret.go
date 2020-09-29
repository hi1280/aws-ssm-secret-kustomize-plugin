package main

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"sigs.k8s.io/kustomize/api/kv"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type plugin struct {
	h                  *resmap.PluginHelpers
	types.ObjectMeta   `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Envs               []string `json:"envs,omitempty" yaml:"envs,omitempty"`
	AwsRegion          string   `json:"region,omitempty" yaml:"region,omitempty"`
	AwsAccessKeyID     string   `json:"awsAccessKeyId,omitempty" yaml:"awsAccessKeyId,omitempty"`
	AwsSecretAccessKey string   `json:"awsSecretAccessKey,omitempty" yaml:"awsSecretAccessKey,omitempty"`
	AwsSessionToken    string   `json:"awsSessionToken,omitempty" yaml:"awsSessionToken,omitempty"`
}

//KustomizePlugin noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(h *resmap.PluginHelpers, c []byte) error {
	p.h = h
	return yaml.Unmarshal(c, p)
}

func (p *plugin) Generate() (resmap.ResMap, error) {
	args := types.SecretArgs{}
	args.Name = p.Name
	args.Namespace = p.Namespace
	cfg := &aws.Config{}
	if p.AwsRegion != "" {
		cfg.Region = aws.String(p.AwsRegion)
	}

	if p.AwsAccessKeyID != "" && p.AwsSecretAccessKey != "" {
		staticCreds := credentials.NewStaticCredentials(p.AwsAccessKeyID, p.AwsSecretAccessKey, p.AwsSessionToken)
		cfg.WithCredentials(staticCreds)
	}

	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}

	svc := ssm.New(sess)

	for _, e := range p.Envs {
		env := strings.Split(e, "=")
		getParamsInput := &ssm.GetParameterInput{
			Name:           aws.String(env[1]),
			WithDecryption: aws.Bool(true),
		}
		resp, err := svc.GetParameter(getParamsInput)
		if err != nil {
			return nil, err
		}
		args.LiteralSources = append(args.LiteralSources, env[0]+"="+*resp.Parameter.Value)
	}

	return p.h.ResmapFactory().FromSecretArgs(
		kv.NewLoader(p.h.Loader(), p.h.Validator()), args)
}

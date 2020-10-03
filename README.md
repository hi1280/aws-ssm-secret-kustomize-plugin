# AWS SSM Secret Kustomize Plugin
Kustomize Plugin to generate Secret using AWS Systems Manager Parameter Store.

## Overview
AWS SSM Secret Kustomize Plugin is a Go plugin for Kustomize that uses the AWS Systems Manager Parameter Store to generate Secrets.
It is used to securely add secrets in Kubernetes.

## Background
When using GitOps, Secret needs to be changed to include rolling update to Pods. kustomize generator plugin supports this.  
It's easy to secure it by using an external AWS secret management system.

## Example

### Requirements
configure your aws credentials as follows.  
https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html#cli-configure-quickstart-config

The following IAM policy allows a user to access parameters.
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ssm:PutParameter",
                "ssm:GetParameter"
            ],
            "Resource": "*"
        }
    ]
}
```

add a secret data.
```sh
aws --region ap-northeast-1 ssm put-parameter --name "/hello-service/password" --type "String" --value "1234"
```

### Setup
When using a kustomize plugin, it must be installed by compiling the kustomize.
```sh
tmpGoPath=$(mktemp -d)
GOPATH=$tmpGoPath go get sigs.k8s.io/kustomize/kustomize/v3
PLUGIN_ROOT=$tmpGoPath/kustomize/plugin
apiVersion=hi1280.com/v1
kind=AwsSsmSecret
lKind=$(echo $kind | awk '{print tolower($0)}')
mkdir -p $PLUGIN_ROOT/${apiVersion}
cd $PLUGIN_ROOT/${apiVersion}
git clone https://github.com/hi1280/aws-ssm-secret-kustomize-plugin.git $lKind
MY_PLUGIN_DIR=$PLUGIN_ROOT/${apiVersion}/${lKind}
cd $MY_PLUGIN_DIR
GOPATH=$tmpGoPath go build -buildmode plugin -o ${kind}.so ${kind}.go
```

### Usage
Build using AWS SSM Secret Kustomize Plugin
```
cd $MY_PLUGIN_DIR
XDG_CONFIG_HOME=$tmpGoPath $tmpGoPath/bin/kustomize build --enable_alpha_plugins example
```

The following is an overview of how to set up the Kustomize Plugin.  
https://kubernetes-sigs.github.io/kustomize/guides/plugins/gopluginguidedexample/

## Argo CD Integration

### Enable Kustomize Plugins via Argo CD ConfigMap

To provide build options to `kustomize build` add a property to the ArgoCD CM under data.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    app.kubernetes.io/name: argocd-cm
    app.kubernetes.io/part-of: argocd
  name: argocd-cm
data:
  kustomize.buildOptions: "--enable_alpha_plugins"
```

### Adding AWS SSM Secret Kustomize Plugin via Volume Mounts

Copy the plugin and kustomize to the repo-server container.  
The following needs to be added to the repo-server manifest.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-repo-server
spec:
  template:
    spec:
      initContainers:
        - name: install-ksops
          image: hi1280/aws-ssm-secret-kustomize-plugin:0.0.1
          command: ["/bin/sh", "-c"]
          args:
            - mv AwsSsmSecret.so /custom-tools/;
              mv /usr/local/bin/kustomize /custom-tools/;
          volumeMounts:
            - mountPath: /custom-tools
              name: custom-tools
      containers:
      - name: argocd-repo-server
        volumeMounts:
        - mountPath: /usr/local/bin/kustomize
          name: custom-tools
          subPath: kustomize
        - mountPath: /.config/kustomize/plugin/hi1280.com/v1/awsssmsecret/AwsSsmSecret.so
          name: custom-tools
          subPath: AwsSsmSecret.so
        env:
          - name: XDG_CONFIG_HOME
            value: /.config
      volumes:
      - name: custom-tools
        emptyDir: {}
```

The following is an overview of how to add tools to the Argo CD.  
https://argoproj.github.io/argo-cd/operator-manual/custom_tools/#adding-tools-via-volume-mounts
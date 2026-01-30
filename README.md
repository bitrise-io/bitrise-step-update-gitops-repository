![Bitrise Badge](https://app.bitrise.io/app/22ca6e807256cbff/status.svg?token=iVhvX_F9mXcYXmBM-qDpng&branch=master)

# Update GitOps repository

Updates files of a GitOps repository either by pushing changes directly to a
given folder of a given branch or by opening a pull request to it.
URL of the pull request is exposed as an output in the latter case.
Updated files are go templates rendered by substituting given values.
A Github username and Personal Access Token must be provided with access to the repository.

## Usage

The step uses the `deployments` input to configure one or more deployment paths to update in a single commit. Each deployment specifies a path in the target repository and the values to apply.

```yaml
inputs:
- templates_folder_path: deployments/helm/service
- deploy_repository_url: https://github.com/bitrise-io/instance-management-deployments.git
- deploy_branch: production
- pull_request: false
- deploy_user: "$DEPLOY_USER"
- commit_message: "${BITRISE_GIT_MESSAGE}"
- deployments: |
    - path: package-foo
      values:
        repository: ${GAR_HOST}/${GAR_PROJECT_PRODUCTION}/${GAR_REPOSITORY}/package-foo
        tag: ${RELEASE_TAG}
    - path: package-bar
      values:
        repository: ${GAR_HOST}/${GAR_PROJECT_PRODUCTION}/${GAR_REPOSITORY}/package-bar
        tag: ${RELEASE_TAG}
```

## Replacer mode

There are situations when simple templating is not sufficient (e.g.: 100 LOC config files). Replacer mode solves this issue by replacing values that match the provided key-delimiter combinations.

### Example

We have a long config file that has the following line:
`us.gcr.io/ip-kubernetes-dev/hello-world-service:tag1`
We would like to replace `tag1` by `tag2` without creating a template.
We can use the step in "Replacer mode", by defining `us.gcr.io/ip-kubernetes-dev/hello-world-service` as `key`, and `:` as `delimiter`. The step will match & replace `tag1` based on the provided key-delimiter combination.

When `replacer_mode` is enabled, `values` expects input a bit differently. Each deployment must also specify the `files` to scan:

```yaml
inputs:
  - deploy_repository_url: $DEPLOY_REPO_URL
  - pull_request: true
  - deploy_user: $DEPLOY_USER
  - deploy_branch: $BRANCH
  - replacer_mode: true
  - delimiter: ":"
  - deployments: |
      - path: service1
        files:
          - config.yaml
        values:
          us.gcr.io/ip-kubernetes-dev/service1: tag2
      - path: service2
        files:
          - config.yaml
        values:
          us.gcr.io/ip-kubernetes-dev/service2: tag3
```

### Caveats

In Replacer mode, the step matches values up until the first comma, exclamation mark, single quote or double quote.

# Development

## How to test this Step

0. Clone this repo
1. Set up .bitrise.secrets.yml file with the following content:

```yaml
envs:
  - DEPLOY_TOKEN: YOUR PAT
  - MY_STEPLIB_REPO_FORK_GIT_URL: YOUR FORK HTTP URL OF THE STEPLIB REPO
```

2. Export these variables in your terminal 
   - `$DEPLOY_USER`: your GH username
   - `$DEPLOY_REPO_URL`: Where to test commit, e.g. https://github.com/bitrise-io/sandbox-deployments.git
   - `$DEPLOY_PATH`: An existing folder in that repo, e.g. zsolt-test
3. Run `bitrise run test` which will open a PR
4. Confirm the PR is opened but close it

## How to release

1. Merge the PR
1. Create a new release / tag
1. Fork the steplib https://github.com/bitrise-io/bitrise-steplib
1. Set YOUR fork steplib URL in the secrets file (see above)
1. Export the step version `$BITRISE_STEP_VERSION` in your terminal, *without* the prefix `v`
1. Run `bitrise run share-this-step`
1. Go to your forked steplib and create a PR
1. Get someone to review your PR and merge it

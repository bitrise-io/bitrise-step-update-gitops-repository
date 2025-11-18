![Bitrise Badge](https://app.bitrise.io/app/22ca6e807256cbff/status.svg?token=iVhvX_F9mXcYXmBM-qDpng&branch=master)

# Update GitOps repository

## Default mode
Updates files of a GitOps repository either by pushing changes directly to a
given folder of a given branch or by opening a pull request to it.
URL of the pull request is exposed as an output in the latter case.
Updated files are go templates rendered by substituting given values.
A Github username and Personal Access Token must be provided with access to the repository.

## Replacer mode
There are situation when simple templating is not sufficient (e.g.: 100 LOC config files). Replacer mode solves this issue by replacing values that match the provided key-delimiter combinations. 

### Example
Example:
We have a long config file that has the following line:
`us.gcr.io/ip-kubernetes-dev/hello-world-service:tag1`
We would like to replace `tag1` by `tag2` without creating a template.
We can use the step in "Replacer mode", by defining `us.gcr.io/ip-kubernetes-dev/hello-world-service` as `key`, and `:` as `delimiter`. The step will match & replace `tag1` based on the provided key-delimiter combination.

When `replacer_mode` is enabled, `values` expects input a bit differently. To achieve the desired outcome above of the above example, one could use the following step inputs:
```yaml
inputs:
  - deploy_repository_url: $DEPLOY_REPO_URL
  - deploy_path: $DEPLOY_PATH
  - pull_request: true
  - deploy_user: $DEPLOY_USER
  - deploy_branch: $BRANCH
  - replacer_mode: true
  - delimiter: ":"
  - files: 
      - example_config_file.yaml
  - values: |
      us.gcr.io/ip-kubernetes-dev/hello-world-service: tag2
```
In this case the step will look for maches in `example_config_file.yaml`, where the full path is `$DEPLOY_PATH/example_config_file.yaml`.

### Caveats
In Replacer mode, the step matches values up until the first comma, exclamation mark, single quote or double quote.

## Regex replacer mode

Regex replacer mode is similar to replacer mode, but instead of looking for exact key matches and delimiters,
it uses regular expressions to find the values to be replaced. It also enables multi-line matching, so it can be used
to match specific parts of the file when there are similar lines.

### Example

Having a section in `globals.yaml` like this:
```yaml
...
    bitrise_cli_checksums:
      macos_arm: 6f7aa1fff4d11f8e45bc106b99b780472155f7b5e70d3f4aca938ab197b54a7c
      macos_amd: a6fd075c26e34fde039412b5cc0691a27a7a917c13ec443a1765a88f975eaba9
      linux_arm: undefined
      linux_amd: c6b5f66dc1deb054cbc81f088eb8205731b98a53b7ac72d8b9135a34acbd1e83
    bitrise_agent_version: 2.52.5
    bitrise_agent_checksums:
      macos_arm: 7e525d446296898bfcaf324fbe090b531479e47a1c9f4e376264fe447e624503
      macos_amd: 0d9425d932b5d62bc99f658f791224d4c620bbc5ce1b72681535a8bdddcd52cd
      linux_arm: undefined
      linux_amd: b531a439264ad4f7706c8e5a779856020e52a5339a0373575920f34c26012f38
    bitrise_agent_version_new: 2.61.3
    bitrise_agent_checksums_new:
      macos_arm: a418a7ec5e493ecfa01d89ff983bba1f47d61074a4a84c739dc3acfa9d34b80b
      macos_amd: a48d1c4b1c64dbfce2a8f609d21670fd80ba8217858e17aba040770e686eef6f
      linux_arm: undefined
      linux_amd: a63747402a8017cb5ebb0a090fa59d53c9c48ed6da5896d9611bdbaefab39116
...
```

Running the step with the following inputs:
```yaml
inputs:
  - deploy_repository_url: $DEPLOY_REPO_URL
  - deploy_path: $DEPLOY_PATH
  - pull_request: true
  - deploy_user: $DEPLOY_USER
  - deploy_branch: production
  - regex_replacer_mode: true
  - files:
    - globals.yaml
  - match_regex: |2 # Indented by two spaces, the rest is part of the value
          bitrise_agent_checksums_new:
            macos_arm: \w+
            macos_amd: \w+
            linux_arm: \w+
            linux_amd: \w+
  - replace_to: |2 # Indented by two spaces, the rest is part of the value
          bitrise_agent_checksums_new:
            macos_arm: checksum-1
            macos_amd: checksum-2
            linux_arm: checksum-3
            linux_amd: checksum-4
```

It'll only match the checksums under the `bitrise_agent_checksums_new` section, and the resulting `globals.yaml` will be:
```yaml
...
    bitrise_cli_checksums:
      macos_arm: 6f7aa1fff4d11f8e45bc106b99b780472155f7b5e70d3f4aca938ab197b54a7c
      macos_amd: a6fd075c26e34fde039412b5cc0691a27a7a917c13ec443a1765a88f975eaba9
      linux_arm: undefined
      linux_amd: c6b5f66dc1deb054cbc81f088eb8205731b98a53b7ac72d8b9135a34acbd1e83
    bitrise_agent_version: 2.52.5
    bitrise_agent_checksums:
      macos_arm: 7e525d446296898bfcaf324fbe090b531479e47a1c9f4e376264fe447e624503
      macos_amd: 0d9425d932b5d62bc99f658f791224d4c620bbc5ce1b72681535a8bdddcd52cd
      linux_arm: undefined
      linux_amd: b531a439264ad4f7706c8e5a779856020e52a5339a0373575920f34c26012f38
    bitrise_agent_version_new: 2.61.3
    bitrise_agent_checksums_new:
      macos_arm: checksum-1
      macos_amd: checksum-2
      linux_arm: checksum-3
      linux_amd: checksum-4
...
```

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

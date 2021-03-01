![Bitrise Badge](https://app.bitrise.io/app/22ca6e807256cbff/status.svg?token=iVhvX_F9mXcYXmBM-qDpng&branch=master)

# Update GitOps repository

Updates files of a GitOps repository either by pushing changes directly to a
given folder of a given branch or by opening a pull request to it.
URL of the pull request is exposed as an output in the latter case.
Updated files are go templates rendered by substituting given values.
A Github username and Personal Access Token must be provided with access to the repository.

## How to use this Step

Can be run directly with the [bitrise CLI](https://github.com/bitrise-io/bitrise),
just `git clone` this repository, `cd` into it's folder in your Terminal/Command Line
and call `bitrise run test`.

*Check the `bitrise.yml` file for required inputs which have to be
added to your `.bitrise.secrets.yml` file!*

Step by step:

1. Open up your Terminal / Command Line
2. `git clone` the repository
3. `cd` into the directory of the step (the one you just `git clone`d)
5. Create a `.bitrise.secrets.yml` file in the same directory of `bitrise.yml`
   (the `.bitrise.secrets.yml` is a git ignored file, you can store your secrets in it)
6. Check the `bitrise.yml` file for any secret you should set in `.bitrise.secrets.yml`
  * Best practice is to mark these options with something like `# define these in your .bitrise.secrets.yml`, in the `app:envs` section.
7. Check the `bitrise.yml` file for any non-sensitive environment variables, export them in your CLI.
8. Once you have all the required environment variables and secret parameters in your `.bitrise.secrets.yml` you can just run this step with the [bitrise CLI](https://github.com/bitrise-io/bitrise): `bitrise run test`

An example `.bitrise.secrets.yml` file:

```
envs:
- DEPLOY_TOKEN: my-pat
```

# Multikey Save Cache

[![Step changelog](https://shields.io/github/v/release/bitrise-steplib/bitrise-step-multikey-save-cache?include_prereleases&label=changelog&color=blueviolet)](https://github.com/bitrise-steplib/bitrise-step-multikey-save-cache/releases)

Restores items from the cache based on a list of keys. This Step needs to be used in combination with **Multikey Save Cache** or **Save Cache**.

<details>
<summary>Description</summary>

Restores items from the cache based on a list of keys.

The format of the keys input is the following:
```
KEY1 || KEY1_ALTERNATIVE1 || KEY1_ALTERNATIVE2
KEY2
KEY3 || KEY3_ALTERNATIVE1
```
The number of keys and paths for each are limited to a number of 10. Commas (`,`) and equal signs (`=`) are not allowed in keys. See templates that can be used in the keys below.

Example (somewhat artificial):
```
multikey_0 || multikey_0_fallback
multikey_1
multikey_2 || multikey_2_fallback
```

This Step needs to be used in combination with **Multikey Save Cache** or **Save Cache**.

#### About key-based caching

Key-based caching is a concept where cache archives are saved and restored using a unique cache key. One Bitrise project can have multiple cache archives stored simultaneously, and the **Multikey Restore Cache** step downloads a cache archive associated with the key provided as a Step input. The **Multikey Save Cache** step is responsible for uploading the cache archive with an exact key.

Caches can become outdated across builds when something changes in the project (for example, a dependency gets upgraded to a new version). In this case, a new (unique) cache key is needed to save the new cache contents. This is possible if the cache key is dynamic and changes based on the project state (for example, a checksum of the dependency lockfile is part of the cache key). If you use the same dynamic cache key when restoring the cache, the Step will download the most relevant cache archive available.

Key-based caching is platform-agnostic and can be used to cache anything by carefully selecting the cache key and the files/folders to include in the cache.

#### Templates

The Step requires a string key to use when uploading a cache archive. In order to always download the most relevant cache archive for each build, the cache key input can contain template elements. The **Multikey Restore Cache** step evaluates the key template at runtime and the final key value can change based on the build environment or files in the repo. Similarly, the **Multikey Save Cache** step also uses templates to compute a unique cache key when uploading a cache archive.

The following variables are supported in the **Cache key** input:

- `cache-key-{{ .Branch }}`: Current git branch the build runs on
- `cache-key-{{ .CommitHash }}`: SHA-256 hash of the git commit the build runs on
- `cache-key-{{ .Workflow }}`: Current Bitrise workflow name (eg. `primary`)
- `{{ .Arch }}-cache-key`: Current CPU architecture (`amd64` or `arm64`)
- `{{ .OS }}-cache-key`: Current operating system (`linux` or `darwin`)

Functions available in a template:

`checksum`: This function takes one or more file paths and computes the SHA256 [checksum](https://en.wikipedia.org/wiki/Checksum) of the file contents. This is useful for creating unique cache keys based on files that describe content to cache.

Examples of using `checksum`:
- `cache-key-{{ checksum "package-lock.json" }}`
- `cache-key-{{ checksum "**/Package.resolved" }}`
- `cache-key-{{ checksum "**/*.gradle*" "gradle.properties" }}`

`getenv`: This function returns the value of an environment variable or an empty string if the variable is not defined.

Examples of `getenv`:
- `cache-key-{{ getenv "PR" }}`
- `cache-key-{{ getenv "BITRISEIO_PIPELINE_ID" }}`

#### Key matching

The most straightforward use case is when both the **Multikey Save Cache** and **Multikey Restore Cache** Steps use the same exact key to transfer cache between builds. Stored cache archives are scoped to the Bitrise project. Builds can restore caches saved by any previous Workflow run on any Bitrise Stack.

The **Multikey Restore Cache** Step can define multiple keys as fallbacks when there is no match for the first cache key.

#### Skip saving the cache

The Step can decide to skip saving a new cache entry to avoid unnecessary work. This happens when there is a previously restored cache in the same workflow and the new cache would have the same contents as the one restored. Make sure to use unique cache keys with a checksum, and enable the **Unique cache key** input for the most optimal execution.

#### Related steps

- [Save Cache](https://github.com/bitrise-steplib/bitrise-step-save-cache/)
- [Restore Cache](https://github.com/bitrise-steplib/bitrise-step-restore-cache/)
- [Multikey Save Cache](https://github.com/bitrise-steplib/bitrise-step-multikey-save-cache/)

</details>

## 🧩 Get started

Add this step directly to your workflow in the [Bitrise Workflow Editor](https://devcenter.bitrise.io/en/steps-and-workflows/introduction-to-steps/adding-steps-to-a-workflow.html).

You can also run this step directly with [Bitrise CLI](https://github.com/bitrise-io/bitrise).

### Examples

Check out [Workflow Recipes](https://github.com/bitrise-io/workflow-recipes?tab=readme-ov-file#-caching) for platform-specific examples!

#### Skip saving the cache in PR builds (only restore)

```yaml
steps:
- multikey-restore-cache@1:
    inputs:
    - keys: |-
        node-modules-{{ checksum "package-lock.json" }}

# Build steps

- multikey-save-cache@1:
    run_if: ".IsCI | and (not .IsPR)" # Condition that is false in PR builds
    inputs:
    - key_path_pairs: |-
          node-modules-{{ checksum "package-lock.json" }} = node_modules
```

#### Separate caches for each OS and architecture

Cache is not guaranteed to work across different Bitrise Stacks (different OS or same OS but different CPU architecture). If a Workflow runs on different stacks, it's a good idea to include the OS and architecture in the **Cache key** input:

```yaml
steps:
- multikey-restore-cache@1:
    inputs:
    - keys: |-
        {{ .OS }}-{{ .Arch }}-npm-cache-{{ checksum "package-lock.json" }} = node_modules
```

#### Multiple independent caches

You can add multiple instances of this Step to a Workflow:

```yaml
steps:
- multikey-restore-cache@1:
    title: Save cache
    inputs:
    - keys: |-
        node-modules-{{ checksum "package-lock.json" }} = node_modules
        pip-packages-{{ checksum "requirements.txt" }} = venv/
```


## ⚙️ Configuration

<details>
<summary>Inputs</summary>

| Key | Description | Flags | Default |
| --- | --- | --- | --- |
| `keys` | Keys used to restore cache archives. The keys support template elements for creating dynamic cache keys. These dynamic keys change the final key value based on the build environment or files in the repo in order to create new cache archives. See the Step description for more details and examples. The maximum length of a key is 512 characters (longer keys get truncated). Commas (`,`) and equal signs (`=`) are not allowed in keys. | required |  |
| `verbose` | Enable logging additional information for troubleshooting | required | `false` |
| `retries` |  Number of retries to attempt when downloading a cache archive fails. The value 0 means no retries are attempted. |  | `3` |
</details>

<details>
<summary>Outputs</summary>

| Key | Description |
| --- | --- |
| `BITRISE_CACHE_HIT` | Indicates if a cache entry was restored. Possible values: `exact`, `partial`, `false` |
</details>

## 🙋 Contributing

We welcome [pull requests](https://github.com/bitrise-steplib/bitrise-step-multikey-save-cache/pulls) and [issues](https://github.com/bitrise-steplib/bitrise-step-multikey-save-cache/issues) against this repository.

For pull requests, work on your changes in a forked repository and use the Bitrise CLI to [run step tests locally](https://devcenter.bitrise.io/en/bitrise-cli/running-your-first-local-build-with-the-cli.html).

**Note:** this step's end-to-end tests (defined in `e2e/bitrise.yml`) are working with secrets which are intentionally not stored in this repo. External contributors won't be able to run those tests. Don't worry, if you open a PR with your contribution, we will help with running tests and make sure that they pass.


Learn more about developing steps:

- [Create your own step](https://devcenter.bitrise.io/en/steps-and-workflows/developing-your-own-bitrise-step.html)

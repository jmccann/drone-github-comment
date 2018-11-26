The Github Comment plugin allows you to add/update comments to issues/PRs from your Drone pipeline.

In it's simplest form, add a comment to the PR triggering the build:

```yaml
pipeline:
  github-comment:
    when:
      event: pull_request
    image: jmccann/drone-github-comment:1
    message: Hello World!
```

You can also have only a single comment added and subsequently updated in later
builds in a PR:

```diff
pipeline:
  github-comment:
    when:
      event: pull_request
    image: jmccann/drone-github-comment:1
    message: Hello World!
+   update: true
```

You can generate fancy comments to a file and have it read in:

```diff
pipeline:
  generate-comment:
    image: alpine:3.7
    commands:
      - echo "Hello World from File!" > comment.file

  github-comment:
    when:
      event: pull_request
    image: jmccann/drone-github-comment:1
-   message: Hello World!
+   message_file: comment.file
```

# Parameter Reference

#### `key`
Override generated comment key to use when updating existing comments.
For PRs, if this is not provided, it's generated automatically.

#### `message`
The message to post.

#### `message_file`
Path to file to read for message to post.

#### `update`
Update existing comment based on `key`. Defaults to `false`.

#### `base_url`
GitHub Base API Url. Example: `https://git.target.com/api/v3`. Defaults to `https://api.github.com`.

#### `api_key`
GitHub API Key.

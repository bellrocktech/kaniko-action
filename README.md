# kaniko-action
Build &amp; push container images with Kaniko

## Warning

This is new/unstable, input values may change. This is untested with public or
other private registries hosted on either GitHub/GitLab or GCP, AWS etc...

## Usage

Use the following example as a template for your own workflow, please refer
to the input parameters for more information.


## Single image build/push

```yaml
      - name: Kaniko builder
        id: kaniko
        uses: bellrocktech/kaniko-action@main
        with:
          registry: reg.example.com
          image: repo/image
          tag: my-tag
          tag_with_latest: true
          path: ./context/path
          username: ${{ secrets.REG_USERNAME }}
          password: ${{ secrets.REG_PASSWORD }}
          cache: true
          cache_url: reg.example.com/cache/image
 ```

## Multiple image build/push

This action is compatible with docker/metadata-action, which allows you to
generate multiple tags.

```yaml

      - name: image metadata
        id: meta
        uses: docker/metadata-action@v4.0.1
        with:
          images: repo/image
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern=v{{version}}
            type=semver,pattern=v{{major}}.{{minor}}
            type=semver,pattern=v{{major}}
            type=sha,prefix=${{ steps.branch-name.outputs.short_ref }}-
            
      - name: Kaniko builder
        id: kaniko
        uses: bellrocktech/kaniko-action@main
        with:
          registry: reg.example.com
          images: ${{ steps.meta.outputs.tags }}
          tag_with_latest: true
          path: ./context/path
          username: ${{ secrets.REG_USERNAME }}
          password: ${{ secrets.REG_PASSWORD }}
          cache: true
          cache_url: reg.example.com/cache/image
 ```


## Inputs

You can either use the `image` and `tag` inputs, OR the `images` input.

For compatibility with docker/metadata-action, you can use the `images` input and
the expected format is a line separated list of `image:tag` pairs.

```yaml
images: |
  repo/image:tag1
  repo/image:tag1
`````

| Name              | Required | Description                                    |
|-------------------|----------|------------------------------------------------|
| `registry`        | true     | The registry to push the image to              |
| `username`        | true     | registry username                              |
| `password`        | true     | registry password                              |
| `images`          | false    | The image names and tags to push to            |
| `image`           | false    | The image name to push to                      |
| `tag`             | false    | The tag to push to                             |
| `tag_with_latest` | false    | Tag the image with latest                      |
| `path`            | false    | build context path                             |
| `dockerfile`      | false    | dockerfile path/name                           |
| `cache`           | false    | enable caching                                 |
| `cache_url`       | false    | cache url                                      |
| `extra_args`      | false    | extra args to pass to kaniko                   |
| `build_args`      | false    | build args to pass to kaniko: FOO=bar,BAR=baz  |
| `target`          | false    | target to build from in multi stage dockerfile |

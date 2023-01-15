# cog to Banan.dev Adapter

If you ever wanted to run something from [Replicate](https://replicate.com/) on your own [Banana.dev](https://banana.dev/) infra, thanks to Replicate's great open model all it needs is a small tool that converts the Cog API format to Banana's.

This repo is that adapter. To use it, create your own repo with a Dockerfile that looks like the following while replaceing user, model, and version hash with what your want:

```Dockerfile
FROM r8.im/{USER}/{MODEL}@sha256:{VERSION_HASH} # e.g. r8.im/nitrosocke/redshift-diffusion@sha256:b78a34f0ec6d21d22ae3b10afd52b219cec65f63362e69e81e4dce07a8154ef8
RUN mkdir /adapter && cd /adapter && wget https://github.com/Hades32/banana-cog-adapter/releases/download/v0.0.7/cog-adapter && chmod +x cog-adapter
ENTRYPOINT ["/adapter/cog-adapter"]
CMD ["python","-m","cog.server.http"]
```

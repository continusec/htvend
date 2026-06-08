load("@gazelle//:deps.bzl", "go_repository")

def go_dependencies():
    go_repository(
        name = "cat_dario_mergo",
        importpath = "dario.cat/mergo",
        sum = "h1:85+piFYR1tMbRrLcDwR18y4UKJ3aH1Tbzi24VRW1TK8=",
        version = "v1.0.2",
    )
    go_repository(
        name = "com_github_acarl005_stripansi",
        importpath = "github.com/acarl005/stripansi",
        sum = "h1:licZJFw2RwpHMqeKTCYkitsPqHNxTmd4SNR5r94FGM8=",
        version = "v0.0.0-20180116102854-5a71ef0e047d",
    )
    go_repository(
        name = "com_github_adrg_xdg",
        importpath = "github.com/adrg/xdg",
        sum = "h1:xRnxJXne7+oWDatRhR1JLnvuccuIeCoBu2rtuLqQB78=",
        version = "v0.5.3",
    )
    go_repository(
        name = "com_github_alecthomas_kingpin_v2",
        importpath = "github.com/alecthomas/kingpin/v2",
        sum = "h1:f48lwail6p8zpO1bC4TxtqACaGqHYA22qkHjHpqDjYY=",
        version = "v2.4.0",
    )
    go_repository(
        name = "com_github_alecthomas_units",
        importpath = "github.com/alecthomas/units",
        sum = "h1:s6gZFSlWYmbqAuRjVTiNNhvNRfY2Wxp9nhfyel4rklc=",
        version = "v0.0.0-20211218093645-b94a6e3cc137",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2",
        importpath = "github.com/aws/aws-sdk-go-v2",
        sum = "h1:DIKX2c31ekm9RA2D9FBj1EWXx++9AdAqRw+e78Tq2Ck=",
        version = "v1.41.12",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_aws_protocol_eventstream",
        importpath = "github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream",
        sum = "h1:p1BBrg/Hhp6uK7zpejeI8QFXHJeC/mynzi04Sl03k9g=",
        version = "v1.7.13",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_config",
        importpath = "github.com/aws/aws-sdk-go-v2/config",
        sum = "h1:PYDobtcsJXK6bQe9I8RQk6s19Bz3xa3xRU08Hy1Em3Y=",
        version = "v1.32.23",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_credentials",
        importpath = "github.com/aws/aws-sdk-go-v2/credentials",
        sum = "h1:SHfH6wyPsEgG7fVsi5rQxWEt7tuIcN2PGhb1mTFv6tE=",
        version = "v1.19.22",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_feature_ec2_imds",
        importpath = "github.com/aws/aws-sdk-go-v2/feature/ec2/imds",
        sum = "h1:b+kcDejJrXc30zU/w8Tc9klISwaO5wh+6T0sMBdDoHM=",
        version = "v1.18.28",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_internal_configsources",
        importpath = "github.com/aws/aws-sdk-go-v2/internal/configsources",
        sum = "h1:Xf2j7NdVcUKomlZ4iihOP4AZ3Fzlr8h4yKpXeP+OFPg=",
        version = "v1.4.28",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_internal_endpoints_v2",
        importpath = "github.com/aws/aws-sdk-go-v2/internal/endpoints/v2",
        sum = "h1:KqIfN9kpkKkcBqBbNpNGTIrXO6ExTUvFKvXkC+YAzVo=",
        version = "v2.7.28",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_internal_v4a",
        importpath = "github.com/aws/aws-sdk-go-v2/internal/v4a",
        sum = "h1:VkE9FuzTQVjBBrnj4+oCdxCLFIz7aqLYKUCjtvxVcOs=",
        version = "v1.4.29",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_internal_accept_encoding",
        importpath = "github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding",
        sum = "h1:ZD2+BSw9vFsNlKYIasSNt3uDbjqqXIBcM13UJv/Lx2k=",
        version = "v1.13.12",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_internal_checksum",
        importpath = "github.com/aws/aws-sdk-go-v2/service/internal/checksum",
        sum = "h1:FsZxbPiVgEHYofziwfylouMki8b1Z7mI4CMU/7bhwBA=",
        version = "v1.9.21",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_internal_presigned_url",
        importpath = "github.com/aws/aws-sdk-go-v2/service/internal/presigned-url",
        sum = "h1:axj4mEDletwKmTm/9jR+DkIMmCfcn5vE4jBMAAN+3Vg=",
        version = "v1.13.28",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_internal_s3shared",
        importpath = "github.com/aws/aws-sdk-go-v2/service/internal/s3shared",
        sum = "h1:li8rTZAAb22g4UsxbjwMdaNVWbgVcDzPqI7nDTI+mF4=",
        version = "v1.19.28",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_s3",
        importpath = "github.com/aws/aws-sdk-go-v2/service/s3",
        sum = "h1:b4ikkRk22T4xYkEgaWc3Voe+3xbt5YbbFhNehOWyUiY=",
        version = "v1.103.2",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_signin",
        importpath = "github.com/aws/aws-sdk-go-v2/service/signin",
        sum = "h1:YcpVyIPLCbiypN6KSphijN5fC7DDjX114SqA7prnnxg=",
        version = "v1.1.4",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_sso",
        importpath = "github.com/aws/aws-sdk-go-v2/service/sso",
        sum = "h1:ySNWu7TPmj5fKFIa1GYvX+Ddxd5ccruqC20aMNuyWDM=",
        version = "v1.31.2",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_ssooidc",
        importpath = "github.com/aws/aws-sdk-go-v2/service/ssooidc",
        sum = "h1:KSzGGqfk39O+WU3OEyYbx6F7sLDQCqxlOJ+2IksfK6U=",
        version = "v1.36.5",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go_v2_service_sts",
        importpath = "github.com/aws/aws-sdk-go-v2/service/sts",
        sum = "h1:RTO7mmGyedgnNmcPh3yQizNfc6GKoV5iqfdJavuf9vw=",
        version = "v1.43.2",
    )
    go_repository(
        name = "com_github_aws_smithy_go",
        importpath = "github.com/aws/smithy-go",
        sum = "h1:4T340VFndXtADGF52gYa1POyL7s9E4Z1OeZ1hCscIw8=",
        version = "v1.27.1",
    )
    go_repository(
        name = "com_github_beorn7_perks",
        importpath = "github.com/beorn7/perks",
        sum = "h1:VlbKKnNfV8bJzeqoa4cOKqO6bYr3WgKZxO8Z16+hsOM=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_burntsushi_toml",
        importpath = "github.com/BurntSushi/toml",
        sum = "h1:W5quZX/G/csjUnuI8SUYlsHs9M38FC7znL0lIO+DvMg=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_cespare_xxhash_v2",
        importpath = "github.com/cespare/xxhash/v2",
        sum = "h1:UL815xU9SqsFlibzuggzjXhog7bL6oX9BbNZnL2UFvs=",
        version = "v2.3.0",
    )
    go_repository(
        name = "com_github_chzyer_readline",
        importpath = "github.com/chzyer/readline",
        sum = "h1:upd/6fQk4src78LMRzh5vItIt361/o4uq553V8B5sGI=",
        version = "v1.5.1",
    )
    go_repository(
        name = "com_github_containerd_cgroups_v3",
        importpath = "github.com/containerd/cgroups/v3",
        sum = "h1:44na7Ud+VwyE7LIoJ8JTNQOa549a8543BmzaJHo6Bzo=",
        version = "v3.0.5",
    )
    go_repository(
        name = "com_github_containerd_errdefs",
        importpath = "github.com/containerd/errdefs",
        sum = "h1:tg5yIfIlQIrxYtu9ajqY42W3lpS19XqdxRQeEwYG8PI=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_containerd_errdefs_pkg",
        importpath = "github.com/containerd/errdefs/pkg",
        sum = "h1:9IKJ06FvyNlexW690DXuQNx2KA2cUJXx151Xdx3ZPPE=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_containerd_log",
        importpath = "github.com/containerd/log",
        sum = "h1:TCJt7ioM2cr/tfR8GPbGf9/VRAX8D2B4PjzCpfX540I=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_containerd_stargz_snapshotter_estargz",
        importpath = "github.com/containerd/stargz-snapshotter/estargz",
        sum = "h1:7evrXtoh1mSbGj/pfRccTampEyKpjpOnS3CyiV1Ebr8=",
        version = "v0.16.3",
    )
    go_repository(
        name = "com_github_containerd_typeurl_v2",
        importpath = "github.com/containerd/typeurl/v2",
        sum = "h1:yNA/94zxWdvYACdYO8zofhrTVuQY73fFU1y++dYSw40=",
        version = "v2.2.3",
    )
    go_repository(
        name = "com_github_containers_image_v5",
        importpath = "github.com/containers/image/v5",
        sum = "h1:GcxYQyAHRF/pLqR4p4RpvKllnNL8mOBn0eZnqJbfTwk=",
        version = "v5.36.2",
    )
    go_repository(
        name = "com_github_containers_libtrust",
        importpath = "github.com/containers/libtrust",
        sum = "h1:Qzk5C6cYglewc+UyGf6lc8Mj2UaPTHy/iF2De0/77CA=",
        version = "v0.0.0-20230121012942-c1716e8a8d01",
    )
    go_repository(
        name = "com_github_containers_ocicrypt",
        importpath = "github.com/containers/ocicrypt",
        sum = "h1:0qIOTT9DoYwcKmxSt8QJt+VzMY18onl9jUXsxpVhSmM=",
        version = "v1.2.1",
    )
    go_repository(
        name = "com_github_containers_storage",
        importpath = "github.com/containers/storage",
        sum = "h1:11Zu68MXsEQGBBd+GadPrHPpWeqjKS8hJDGiAHgIqDs=",
        version = "v1.59.1",
    )
    go_repository(
        name = "com_github_coreos_go_oidc_v3",
        importpath = "github.com/coreos/go-oidc/v3",
        sum = "h1:9ePWwfdwC4QKRlCXsJGou56adA/owXczOzwKdOumLqk=",
        version = "v3.14.1",
    )
    go_repository(
        name = "com_github_creack_pty",
        importpath = "github.com/creack/pty",
        sum = "h1:uDmaGzcdjhF4i/plgjmEsriH11Y0o7RKapEf/LDaM3w=",
        version = "v1.1.9",
    )
    go_repository(
        name = "com_github_cyberphone_json_canonicalization",
        importpath = "github.com/cyberphone/json-canonicalization",
        sum = "h1:uX1JmpONuD549D73r6cgnxyUu18Zb7yHAy5AYU0Pm4Q=",
        version = "v0.0.0-20241213102144-19d51d7fe467",
    )
    go_repository(
        name = "com_github_cyphar_filepath_securejoin",
        importpath = "github.com/cyphar/filepath-securejoin",
        sum = "h1:JyxxyPEaktOD+GAnqIqTf9A8tHyAG22rowi7HkoSU1s=",
        version = "v0.4.1",
    )
    go_repository(
        name = "com_github_danjacques_gofslock",
        importpath = "github.com/danjacques/gofslock",
        sum = "h1:m+Fkk9QEMuV6Z1ithqqYogOHV7Pl6rMKe34NBTJTS/c=",
        version = "v0.0.0-20240212154529-d899e02bfe22",
    )
    go_repository(
        name = "com_github_davecgh_go_spew",
        importpath = "github.com/davecgh/go-spew",
        sum = "h1:U9qPSI2PIWSS1VwoXQT9A3Wy9MM3WgvqSxFWenqJduM=",
        version = "v1.1.2-0.20180830191138-d8f796af33cc",
    )
    go_repository(
        name = "com_github_distribution_reference",
        importpath = "github.com/distribution/reference",
        sum = "h1:0IXCQ5g4/QMHHkarYzh5l+u8T3t73zM5QvfrDyIgxBk=",
        version = "v0.6.0",
    )
    go_repository(
        name = "com_github_docker_cli",
        importpath = "github.com/docker/cli",
        sum = "h1:mOt9fcLE7zaACbxW1GeS65RI67wIJrTnqS3hP2huFsY=",
        version = "v28.3.2+incompatible",
    )
    go_repository(
        name = "com_github_docker_distribution",
        importpath = "github.com/docker/distribution",
        sum = "h1:AtKxIZ36LoNK51+Z6RpzLpddBirtxJnzDrHLEKxTAYk=",
        version = "v2.8.3+incompatible",
    )
    go_repository(
        name = "com_github_docker_docker",
        importpath = "github.com/docker/docker",
        sum = "h1:wn66NJ6pWB1vBZIilP8G3qQPqHy5XymfYn5vsqeA5oA=",
        version = "v28.3.2+incompatible",
    )
    go_repository(
        name = "com_github_docker_docker_credential_helpers",
        importpath = "github.com/docker/docker-credential-helpers",
        sum = "h1:gAm/VtF9wgqJMoxzT3Gj5p4AqIjCBS4wrsOh9yRqcz8=",
        version = "v0.9.3",
    )
    go_repository(
        name = "com_github_docker_go_connections",
        importpath = "github.com/docker/go-connections",
        sum = "h1:USnMq7hx7gwdVZq1L49hLXaFtUdTADjXGp+uj1Br63c=",
        version = "v0.5.0",
    )
    go_repository(
        name = "com_github_docker_go_metrics",
        importpath = "github.com/docker/go-metrics",
        sum = "h1:AgB/0SvBxihN0X8OR4SjsblXkbMvalQ8cjmtKQ2rQV8=",
        version = "v0.0.1",
    )
    go_repository(
        name = "com_github_docker_go_units",
        importpath = "github.com/docker/go-units",
        sum = "h1:69rxXcBk27SvSaaxTtLh/8llcHD8vYHT7WSdRZ/jvr4=",
        version = "v0.5.0",
    )
    go_repository(
        name = "com_github_docker_libtrust",
        importpath = "github.com/docker/libtrust",
        sum = "h1:UhxFibDNY/bfvqU5CAUmr9zpesgbU6SWc8/B4mflAE4=",
        version = "v0.0.0-20160708172513-aabc10ec26b7",
    )
    go_repository(
        name = "com_github_felixge_httpsnoop",
        importpath = "github.com/felixge/httpsnoop",
        sum = "h1:NFTV2Zj1bL4mc9sqWACXbQFVBBg2W3GPvqp8/ESS2Wg=",
        version = "v1.0.4",
    )
    go_repository(
        name = "com_github_gboddin_go_www_authenticate_parser",
        importpath = "github.com/gboddin/go-www-authenticate-parser",
        sum = "h1:JvEO7eltd2aCHF+ABLquTUziO7hzC6G7H3tgENYkDBc=",
        version = "v0.0.0-20230926203616-ec0b649bb077",
    )
    go_repository(
        name = "com_github_go_jose_go_jose_v4",
        importpath = "github.com/go-jose/go-jose/v4",
        sum = "h1:M6T8+mKZl/+fNNuFHvGIzDz7BTLQPIounk/b9dw3AaE=",
        version = "v4.0.5",
    )
    go_repository(
        name = "com_github_go_logr_logr",
        importpath = "github.com/go-logr/logr",
        sum = "h1:CjnDlHq8ikf6E492q6eKboGOC0T8CDaOvkHCIg8idEI=",
        version = "v1.4.3",
    )
    go_repository(
        name = "com_github_go_logr_stdr",
        importpath = "github.com/go-logr/stdr",
        sum = "h1:hSWxHoqTgW2S2qGc0LTAI563KZ5YKYRhT3MFKZMbjag=",
        version = "v1.2.2",
    )
    go_repository(
        name = "com_github_gogo_protobuf",
        importpath = "github.com/gogo/protobuf",
        sum = "h1:Ov1cvc58UF3b5XjBnZv7+opcTcQFZebYjWzi34vdm4Q=",
        version = "v1.3.2",
    )
    go_repository(
        name = "com_github_golang_groupcache",
        importpath = "github.com/golang/groupcache",
        sum = "h1:f+oWsMOmNPc8JmEHVZIycC7hBoQxHH9pNKQORJNozsQ=",
        version = "v0.0.0-20241129210726-2c02b8208cf8",
    )
    go_repository(
        name = "com_github_golang_protobuf",
        importpath = "github.com/golang/protobuf",
        sum = "h1:i7eJL8qZTpSEXOPTxNKhASYpMn+8e5Q6AdndVa1dWek=",
        version = "v1.5.4",
    )
    go_repository(
        name = "com_github_google_go_cmp",
        importpath = "github.com/google/go-cmp",
        sum = "h1:wk8382ETsv4JYUZwIsn6YpYiWiBsYLSJiTsyBybVuN8=",
        version = "v0.7.0",
    )
    go_repository(
        name = "com_github_google_go_containerregistry",
        importpath = "github.com/google/go-containerregistry",
        sum = "h1:oNx7IdTI936V8CQRveCjaxOiegWwvM7kqkbXTpyiovI=",
        version = "v0.20.3",
    )
    go_repository(
        name = "com_github_google_go_intervals",
        importpath = "github.com/google/go-intervals",
        sum = "h1:FGrVEiUnTRKR8yE04qzXYaJMtnIYqobR5QbblK3ixcM=",
        version = "v0.0.2",
    )
    go_repository(
        name = "com_github_google_uuid",
        importpath = "github.com/google/uuid",
        sum = "h1:NIvaJDMOsjHA8n1jAhLSgzrAzy1Hgr+hNrb57e+94F0=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_gorilla_mux",
        importpath = "github.com/gorilla/mux",
        sum = "h1:TuBL49tXwgrFYWhqrNgrUNEY92u81SPhu7sTdzQEiWY=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_github_hashicorp_errwrap",
        importpath = "github.com/hashicorp/errwrap",
        sum = "h1:hLrqtEDnRye3+sgx6z4qVLNuviH3MR5aQ0ykNJa/UYA=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_hashicorp_go_cleanhttp",
        importpath = "github.com/hashicorp/go-cleanhttp",
        sum = "h1:035FKYIWjmULyFRBKPs8TBQoi0x6d9G4xc9neXJWAZQ=",
        version = "v0.5.2",
    )
    go_repository(
        name = "com_github_hashicorp_go_multierror",
        importpath = "github.com/hashicorp/go-multierror",
        sum = "h1:H5DkEtf6CXdFp0N0Em5UCwQpXMWke8IA0+lD48awMYo=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_hashicorp_go_retryablehttp",
        importpath = "github.com/hashicorp/go-retryablehttp",
        sum = "h1:ylXZWnqa7Lhqpk0L1P1LzDtGcCR0rPVUrx/c8Unxc48=",
        version = "v0.7.8",
    )
    go_repository(
        name = "com_github_jessevdk_go_flags",
        importpath = "github.com/jessevdk/go-flags",
        sum = "h1:Cvu5U8UGrLay1rZfv/zP7iLpSHGUZ/Ou68T0iX1bBK4=",
        version = "v1.6.1",
    )
    go_repository(
        name = "com_github_jpillora_backoff",
        importpath = "github.com/jpillora/backoff",
        sum = "h1:uvFg412JmmHBHw7iwprIxkPMI+sGQ4kzOWsMeHnm2EA=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_json_iterator_go",
        importpath = "github.com/json-iterator/go",
        sum = "h1:PV8peI4a0ysnczrg+LtxykD8LfKY9ML6u2jnxaEnrnM=",
        version = "v1.1.12",
    )
    go_repository(
        name = "com_github_julienschmidt_httprouter",
        importpath = "github.com/julienschmidt/httprouter",
        sum = "h1:U0609e9tgbseu3rBINet9P48AI/D3oJs4dN7jwJOQ1U=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_klauspost_compress",
        importpath = "github.com/klauspost/compress",
        sum = "h1:c/Cqfb0r+Yi+JtIEq73FWXVkRonBlf0CRNYc8Zttxdo=",
        version = "v1.18.0",
    )
    go_repository(
        name = "com_github_klauspost_pgzip",
        importpath = "github.com/klauspost/pgzip",
        sum = "h1:8RXeL5crjEUFnR2/Sn6GJNWtSQ3Dk8pq4CL3jvdDyjU=",
        version = "v1.2.6",
    )
    go_repository(
        name = "com_github_kr_pretty",
        importpath = "github.com/kr/pretty",
        sum = "h1:flRD4NNwYAUpkphVc1HcthR4KEIFJ65n8Mw5qdRn3LE=",
        version = "v0.3.1",
    )
    go_repository(
        name = "com_github_kr_text",
        importpath = "github.com/kr/text",
        sum = "h1:5Nx0Ya0ZqY2ygV366QzturHI13Jq95ApcVaJBhpS+AY=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_kylelemons_godebug",
        importpath = "github.com/kylelemons/godebug",
        sum = "h1:RPNrshWIDI6G2gRW9EHilWtl7Z6Sb1BR0xunSBf0SNc=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_letsencrypt_boulder",
        importpath = "github.com/letsencrypt/boulder",
        sum = "h1:2tTW6cDth2TSgRbAhD7yjZzTQmcN25sDRPEeinR51yQ=",
        version = "v0.0.0-20240620165639-de9c06129bec",
    )
    go_repository(
        name = "com_github_manifoldco_promptui",
        importpath = "github.com/manifoldco/promptui",
        sum = "h1:3V4HzJk1TtXW1MTZMP7mdlwbBpIinw3HztaIlYthEiA=",
        version = "v0.9.0",
    )
    go_repository(
        name = "com_github_mattn_go_runewidth",
        importpath = "github.com/mattn/go-runewidth",
        sum = "h1:E5ScNMtiwvlvB5paMFdw9p4kSQzbXFikJ5SQO6TULQc=",
        version = "v0.0.16",
    )
    go_repository(
        name = "com_github_mattn_go_shellwords",
        importpath = "github.com/mattn/go-shellwords",
        sum = "h1:M2zGm7EW6UQJvDeQxo4T51eKPurbeFbe8WtebGE2xrk=",
        version = "v1.0.12",
    )
    go_repository(
        name = "com_github_mattn_go_sqlite3",
        importpath = "github.com/mattn/go-sqlite3",
        sum = "h1:ThEiQrnbtumT+QMknw63Befp/ce/nUPgBPMlRFEum7A=",
        version = "v1.14.28",
    )
    go_repository(
        name = "com_github_microsoft_go_winio",
        importpath = "github.com/Microsoft/go-winio",
        sum = "h1:F2VQgta7ecxGYO8k3ZZz3RS8fVIXVxONVUPlNERoyfY=",
        version = "v0.6.2",
    )
    go_repository(
        name = "com_github_microsoft_hcsshim",
        importpath = "github.com/Microsoft/hcsshim",
        sum = "h1:/BcXOiS6Qi7N9XqUcv27vkIuVOkBEcWstd2pMlWSeaA=",
        version = "v0.13.0",
    )
    go_repository(
        name = "com_github_miekg_pkcs11",
        importpath = "github.com/miekg/pkcs11",
        sum = "h1:Ugu9pdy6vAYku5DEpVWVFPYnzV+bxB+iRdbuFSu7TvU=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_mistifyio_go_zfs_v3",
        importpath = "github.com/mistifyio/go-zfs/v3",
        sum = "h1:YaoXgBePoMA12+S1u/ddkv+QqxcfiZK4prI6HPnkFiU=",
        version = "v3.0.1",
    )
    go_repository(
        name = "com_github_moby_docker_image_spec",
        importpath = "github.com/moby/docker-image-spec",
        sum = "h1:jMKff3w6PgbfSa69GfNg+zN/XLhfXJGnEx3Nl2EsFP0=",
        version = "v1.3.1",
    )
    go_repository(
        name = "com_github_moby_sys_atomicwriter",
        importpath = "github.com/moby/sys/atomicwriter",
        sum = "h1:kw5D/EqkBwsBFi0ss9v1VG3wIkVhzGvLklJ+w3A14Sw=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_moby_sys_capability",
        importpath = "github.com/moby/sys/capability",
        sum = "h1:4D4mI6KlNtWMCM1Z/K0i7RV1FkX+DBDHKVJpCndZoHk=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_moby_sys_mountinfo",
        importpath = "github.com/moby/sys/mountinfo",
        sum = "h1:1shs6aH5s4o5H2zQLn796ADW1wMrIwHsyJ2v9KouLrg=",
        version = "v0.7.2",
    )
    go_repository(
        name = "com_github_moby_sys_user",
        importpath = "github.com/moby/sys/user",
        sum = "h1:jhcMKit7SA80hivmFJcbB1vqmw//wU61Zdui2eQXuMs=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_moby_term",
        importpath = "github.com/moby/term",
        sum = "h1:6qk3FJAFDs6i/q3W/pQ97SX192qKfZgGjCQqfCJkgzQ=",
        version = "v0.5.2",
    )
    go_repository(
        name = "com_github_modern_go_concurrent",
        importpath = "github.com/modern-go/concurrent",
        sum = "h1:TRLaZ9cD/w8PVh93nsPXa1VrQ6jlwL5oN8l14QlcNfg=",
        version = "v0.0.0-20180306012644-bacd9c7ef1dd",
    )
    go_repository(
        name = "com_github_modern_go_reflect2",
        importpath = "github.com/modern-go/reflect2",
        sum = "h1:xBagoLtFs94CBntxluKeaWgTMpvLxC4ur3nMaC9Gz0M=",
        version = "v1.0.2",
    )
    go_repository(
        name = "com_github_munnerz_goautoneg",
        importpath = "github.com/munnerz/goautoneg",
        sum = "h1:C3w9PqII01/Oq1c1nUAm88MOHcQC9l5mIlSMApZMrHA=",
        version = "v0.0.0-20191010083416-a7dc8b61c822",
    )
    go_repository(
        name = "com_github_mwitkow_go_conntrack",
        importpath = "github.com/mwitkow/go-conntrack",
        sum = "h1:KUppIJq7/+SVif2QVs3tOP0zanoHgBEVAwHxUSIzRqU=",
        version = "v0.0.0-20190716064945-2f068394615f",
    )
    go_repository(
        name = "com_github_opencontainers_go_digest",
        importpath = "github.com/opencontainers/go-digest",
        sum = "h1:apOUWs51W5PlhuyGyz9FCeeBIOUDA/6nW8Oi/yOhh5U=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_opencontainers_image_spec",
        importpath = "github.com/opencontainers/image-spec",
        sum = "h1:y0fUlFfIZhPF1W537XOLg0/fcx6zcHCJwooC2xJA040=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_opencontainers_runtime_spec",
        importpath = "github.com/opencontainers/runtime-spec",
        sum = "h1:S4k4ryNgEpxW1dzyqffOmhI1BHYcjzU8lpJfSlR0xww=",
        version = "v1.2.1",
    )
    go_repository(
        name = "com_github_opencontainers_selinux",
        importpath = "github.com/opencontainers/selinux",
        sum = "h1:6n5JV4Cf+4y0KNXW48TLj5DwfXpvWlxXplUkdTrmPb8=",
        version = "v1.12.0",
    )
    go_repository(
        name = "com_github_pkg_errors",
        importpath = "github.com/pkg/errors",
        sum = "h1:FEBLx1zS214owpjy7qsBeixbURkuhQAwrK5UwLGTwt4=",
        version = "v0.9.1",
    )
    go_repository(
        name = "com_github_pmezard_go_difflib",
        importpath = "github.com/pmezard/go-difflib",
        sum = "h1:Jamvg5psRIccs7FGNTlIRMkT8wgtp5eCXdBlqhYGL6U=",
        version = "v1.0.1-0.20181226105442-5d4384ee4fb2",
    )
    go_repository(
        name = "com_github_proglottis_gpgme",
        importpath = "github.com/proglottis/gpgme",
        sum = "h1:3nE7YNA70o2aLjcg63tXMOhPD7bplfE5CBdV+hLAm2M=",
        version = "v0.1.4",
    )
    go_repository(
        name = "com_github_prometheus_client_golang",
        importpath = "github.com/prometheus/client_golang",
        sum = "h1:Je96obch5RDVy3FDMndoUsjAhG5Edi49h0RJWRi/o0o=",
        version = "v1.23.2",
    )
    go_repository(
        name = "com_github_prometheus_client_model",
        importpath = "github.com/prometheus/client_model",
        sum = "h1:oBsgwpGs7iVziMvrGhE53c/GrLUsZdHnqNwqPLxwZyk=",
        version = "v0.6.2",
    )
    go_repository(
        name = "com_github_prometheus_common",
        importpath = "github.com/prometheus/common",
        sum = "h1:h5E0h5/Y8niHc5DlaLlWLArTQI7tMrsfQjHV+d9ZoGs=",
        version = "v0.66.1",
    )
    go_repository(
        name = "com_github_prometheus_procfs",
        importpath = "github.com/prometheus/procfs",
        sum = "h1:hZ15bTNuirocR6u0JZ6BAHHmwS1p8B4P6MRqxtzMyRg=",
        version = "v0.16.1",
    )
    go_repository(
        name = "com_github_rivo_uniseg",
        importpath = "github.com/rivo/uniseg",
        sum = "h1:WUdvkW8uEhrYfLC4ZzdpI2ztxP1I582+49Oc5Mq64VQ=",
        version = "v0.4.7",
    )
    go_repository(
        name = "com_github_rogpeppe_go_internal",
        importpath = "github.com/rogpeppe/go-internal",
        sum = "h1:TMyTOH3F/DB16zRVcYyreMH6GnZZrwQVAoYjRBZyWFQ=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_github_russross_blackfriday",
        importpath = "github.com/russross/blackfriday",
        sum = "h1:KqfZb0pUVN2lYqZUYRddxF4OR8ZMURnJIG5Y3VRLtww=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_santhosh_tekuri_jsonschema_v5",
        importpath = "github.com/santhosh-tekuri/jsonschema/v5",
        sum = "h1:lZUw3E0/J3roVtGQ+SCrUrg3ON6NgVqpn3+iol9aGu4=",
        version = "v5.3.1",
    )
    go_repository(
        name = "com_github_santhosh_tekuri_jsonschema_v6",
        importpath = "github.com/santhosh-tekuri/jsonschema/v6",
        sum = "h1:KRzFb2m7YtdldCEkzs6KqmJw4nqEVZGK7IN2kJkjTuQ=",
        version = "v6.0.2",
    )
    go_repository(
        name = "com_github_secure_systems_lab_go_securesystemslib",
        importpath = "github.com/secure-systems-lab/go-securesystemslib",
        sum = "h1:rf1HIbL64nUpEIZnjLZ3mcNEL9NBPB0iuVjyxvq3LZc=",
        version = "v0.9.0",
    )
    go_repository(
        name = "com_github_segmentio_ksuid",
        importpath = "github.com/segmentio/ksuid",
        sum = "h1:sBo2BdShXjmcugAMwjugoGUdUV0pcxY5mW4xKRn3v4c=",
        version = "v1.0.4",
    )
    go_repository(
        name = "com_github_sergi_go_diff",
        importpath = "github.com/sergi/go-diff",
        sum = "h1:xkr+Oxo4BOQKmkn/B9eMK0g5Kg/983T9DqqPHwYqD+8=",
        version = "v1.3.1",
    )
    go_repository(
        name = "com_github_sigstore_fulcio",
        importpath = "github.com/sigstore/fulcio",
        sum = "h1:XaMYX6TNT+8n7Npe8D94nyZ7/ERjEsNGFC+REdi/wzw=",
        version = "v1.6.6",
    )
    go_repository(
        name = "com_github_sigstore_protobuf_specs",
        importpath = "github.com/sigstore/protobuf-specs",
        sum = "h1:5SsMqZbdkcO/DNHudaxuCUEjj6x29tS2Xby1BxGU7Zc=",
        version = "v0.4.1",
    )
    go_repository(
        name = "com_github_sigstore_sigstore",
        importpath = "github.com/sigstore/sigstore",
        sum = "h1:Wm1LT9yF4LhQdEMy5A2JeGRHTrAWGjT3ubE5JUSrGVU=",
        version = "v1.9.5",
    )
    go_repository(
        name = "com_github_sirupsen_logrus",
        importpath = "github.com/sirupsen/logrus",
        sum = "h1:TsZE7l11zFCLZnZ+teH4Umoq5BhEIfIzfRDZ1Uzql2w=",
        version = "v1.9.4",
    )
    go_repository(
        name = "com_github_skratchdot_open_golang",
        importpath = "github.com/skratchdot/open-golang",
        sum = "h1:JIAuq3EEf9cgbU6AtGPK4CTG3Zf6CKMNqf0MHTggAUA=",
        version = "v0.0.0-20200116055534-eef842397966",
    )
    go_repository(
        name = "com_github_smallstep_pkcs7",
        importpath = "github.com/smallstep/pkcs7",
        sum = "h1:x+rPdt2W088V9Vkjho4KtoggyktZJlMduZAtRHm68LU=",
        version = "v0.1.1",
    )
    go_repository(
        name = "com_github_stefanberger_go_pkcs11uri",
        importpath = "github.com/stefanberger/go-pkcs11uri",
        sum = "h1:pnnLyeX7o/5aX8qUQ69P/mLojDqwda8hFOCBTmP/6hw=",
        version = "v0.0.0-20230803200340-78284954bff6",
    )
    go_repository(
        name = "com_github_stretchr_objx",
        importpath = "github.com/stretchr/objx",
        sum = "h1:xuMeJ0Sdp5ZMRXx/aWO6RZxdr3beISkG5/G/aIRr3pY=",
        version = "v0.5.2",
    )
    go_repository(
        name = "com_github_stretchr_testify",
        importpath = "github.com/stretchr/testify",
        sum = "h1:7s2iGBzp5EwR7/aIZr8ao5+dra3wiQyKjjFuvgVKu7U=",
        version = "v1.11.1",
    )
    go_repository(
        name = "com_github_sylabs_sif_v2",
        importpath = "github.com/sylabs/sif/v2",
        sum = "h1:GZ0b5//AFAqJEChd8wHV/uSKx/l1iuGYwjR8nx+4wPI=",
        version = "v2.21.1",
    )
    go_repository(
        name = "com_github_tchap_go_patricia_v2",
        importpath = "github.com/tchap/go-patricia/v2",
        sum = "h1:xfNEsODumaEcCcY3gI0hYPZ/PcpVv5ju6RMAhgwZDDc=",
        version = "v2.3.3",
    )
    go_repository(
        name = "com_github_titanous_rocacheck",
        importpath = "github.com/titanous/rocacheck",
        sum = "h1:e/5i7d4oYZ+C1wj2THlRK+oAhjeS/TRQwMfkIuet3w0=",
        version = "v0.0.0-20171023193734-afe73141d399",
    )
    go_repository(
        name = "com_github_ulikunitz_xz",
        importpath = "github.com/ulikunitz/xz",
        sum = "h1:37Nm15o69RwBkXM0J6A5OlE67RZTfzUxTj8fB3dfcsc=",
        version = "v0.5.12",
    )
    go_repository(
        name = "com_github_vbatts_tar_split",
        importpath = "github.com/vbatts/tar-split",
        sum = "h1:CqKoORW7BUWBe7UL/iqTVvkTBOF8UvOMKOIZykxnnbo=",
        version = "v0.12.1",
    )
    go_repository(
        name = "com_github_vbauerster_mpb_v8",
        importpath = "github.com/vbauerster/mpb/v8",
        sum = "h1:2uBykSHAYHekE11YvJhKxYmLATKHAGorZwFlyNw4hHM=",
        version = "v8.10.2",
    )
    go_repository(
        name = "com_github_vividcortex_ewma",
        importpath = "github.com/VividCortex/ewma",
        sum = "h1:f58SaIzcDXrSy3kWaHNvuJgJ3Nmz59Zji6XoJR/q1ow=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_xhit_go_str2duration_v2",
        importpath = "github.com/xhit/go-str2duration/v2",
        sum = "h1:lxklc02Drh6ynqX+DdPyp5pCKLUQpRT8bp8Ydu2Bstc=",
        version = "v2.1.0",
    )
    go_repository(
        name = "com_sslmate_software_src_go_pkcs12",
        importpath = "software.sslmate.com/src/go-pkcs12",
        sum = "h1:bxkUPRsvTPNRBZa4M/aSX4PyMOEbq3V8I6hbkG4F4Q8=",
        version = "v0.7.1",
    )
    go_repository(
        name = "in_gopkg_check_v1",
        importpath = "gopkg.in/check.v1",
        sum = "h1:Hei/4ADfdWqJk1ZMxUNpqntNwaWcugrBjAiHlqqRiVk=",
        version = "v1.0.0-20201130134442-10cb98267c6c",
    )
    go_repository(
        name = "in_gopkg_yaml_v3",
        importpath = "gopkg.in/yaml.v3",
        sum = "h1:fxVm/GzAzEWqLHuvctI91KS9hhNmmWOoWu0XTYJS7CA=",
        version = "v3.0.1",
    )
    go_repository(
        name = "in_yaml_go_yaml_v2",
        importpath = "go.yaml.in/yaml/v2",
        sum = "h1:DzmwEr2rDGHl7lsFgAHxmNz/1NlQ7xLIrlN2h5d1eGI=",
        version = "v2.4.2",
    )
    go_repository(
        name = "io_etcd_go_bbolt",
        importpath = "go.etcd.io/bbolt",
        sum = "h1:IrUHp260R8c+zYx/Tm8QZr04CX+qWS5PGfPdevhdm1I=",
        version = "v1.4.2",
    )
    go_repository(
        name = "io_opencensus_go",
        importpath = "go.opencensus.io",
        sum = "h1:y73uSU6J157QMP2kn2r30vwW1A2W2WFwSCGnAVxeaD0=",
        version = "v0.24.0",
    )
    go_repository(
        name = "io_opentelemetry_go_auto_sdk",
        importpath = "go.opentelemetry.io/auto/sdk",
        sum = "h1:cH53jehLUN6UFLY71z+NDOiNJqDdPRaXzTel0sJySYA=",
        version = "v1.1.0",
    )
    go_repository(
        name = "io_opentelemetry_go_contrib_instrumentation_net_http_otelhttp",
        importpath = "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp",
        sum = "h1:sbiXRNDSWJOTobXh5HyQKjq6wUC5tNybqjIqDpAY4CU=",
        version = "v0.60.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel",
        importpath = "go.opentelemetry.io/otel",
        sum = "h1:xKWKPxrxB6OtMCbmMY021CqC45J+3Onta9MqjhnusiQ=",
        version = "v1.35.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_exporters_otlp_otlptrace_otlptracehttp",
        importpath = "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp",
        sum = "h1:xJ2qHD0C1BeYVTLLR9sX12+Qb95kfeD/byKj6Ky1pXg=",
        version = "v1.35.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_metric",
        importpath = "go.opentelemetry.io/otel/metric",
        sum = "h1:0znxYu2SNyuMSQT4Y9WDWej0VpcsxkuklLa4/siN90M=",
        version = "v1.35.0",
    )
    go_repository(
        name = "io_opentelemetry_go_otel_trace",
        importpath = "go.opentelemetry.io/otel/trace",
        sum = "h1:dPpEfJu1sDIqruz7BHFG3c7528f6ddfSWfFDVt/xgMs=",
        version = "v1.35.0",
    )
    go_repository(
        name = "org_golang_google_genproto",
        importpath = "google.golang.org/genproto",
        sum = "h1:KpwkzHKEF7B9Zxg18WzOa7djJ+Ha5DzthMyZYQfEn2A=",
        version = "v0.0.0-20230410155749-daa745c078e1",
    )
    go_repository(
        name = "org_golang_google_genproto_googleapis_api",
        importpath = "google.golang.org/genproto/googleapis/api",
        sum = "h1:p31xT4yrYrSM/G4Sn2+TNUkVhFCbG9y8itM2S6Th950=",
        version = "v0.0.0-20250303144028-a0af3efb3deb",
    )
    go_repository(
        name = "org_golang_google_genproto_googleapis_rpc",
        importpath = "google.golang.org/genproto/googleapis/rpc",
        sum = "h1:iK2jbkWL86DXjEx0qiHcRE9dE4/Ahua5k6V8OWFb//c=",
        version = "v0.0.0-20250313205543-e70fdf4c4cb4",
    )
    go_repository(
        name = "org_golang_google_grpc",
        importpath = "google.golang.org/grpc",
        sum = "h1:TdbGzwb82ty4OusHWepvFWGLgIbNo1/SUynEN0ssqv8=",
        version = "v1.72.2",
    )
    go_repository(
        name = "org_golang_google_protobuf",
        importpath = "google.golang.org/protobuf",
        sum = "h1:xHScyCOEuuwZEc6UtSOvPbAT4zRh0xcNRYekJwfqyMc=",
        version = "v1.36.8",
    )
    go_repository(
        name = "org_golang_x_crypto",
        importpath = "golang.org/x/crypto",
        sum = "h1:jMBrvKuj23MTlT0bQEOBcAE0mjg8mK9RXFhRH6nyF3Q=",
        version = "v0.45.0",
    )
    go_repository(
        name = "org_golang_x_net",
        importpath = "golang.org/x/net",
        sum = "h1:Mx+4dIFzqraBXUugkia1OOvlD6LemFo1ALMHjrXDOhY=",
        version = "v0.47.0",
    )
    go_repository(
        name = "org_golang_x_oauth2",
        importpath = "golang.org/x/oauth2",
        sum = "h1:dnDm7JmhM45NNpd8FDDeLhK6FwqbOf4MLCM9zb1BOHI=",
        version = "v0.30.0",
    )
    go_repository(
        name = "org_golang_x_sync",
        importpath = "golang.org/x/sync",
        sum = "h1:ycBJEhp9p4vXvUZNszeOq0kGTPghopOL8q0fq3vstxw=",
        version = "v0.16.0",
    )
    go_repository(
        name = "org_golang_x_sys",
        importpath = "golang.org/x/sys",
        sum = "h1:3yZWxaJjBmCWXqhN1qh02AkOnCQ1poK6oF+a7xWL6Gc=",
        version = "v0.38.0",
    )
    go_repository(
        name = "org_golang_x_term",
        importpath = "golang.org/x/term",
        sum = "h1:8EGAD0qCmHYZg6J17DvsMy9/wJ7/D/4pV/wfnld5lTU=",
        version = "v0.37.0",
    )
    go_repository(
        name = "org_golang_x_text",
        importpath = "golang.org/x/text",
        sum = "h1:aC8ghyu4JhP8VojJ2lEHBnochRno1sgL6nEi9WGFGMM=",
        version = "v0.31.0",
    )
    go_repository(
        name = "org_golang_x_time",
        importpath = "golang.org/x/time",
        sum = "h1:/bpjEDfN9tkoN/ryeYHnv5hcMlc8ncjMcM4XBk5NWV0=",
        version = "v0.11.0",
    )
    go_repository(
        name = "org_golang_x_xerrors",
        importpath = "golang.org/x/xerrors",
        sum = "h1:E7g+9GITq07hpfrRu66IVDexMakfv52eLZ2CXBWiKr4=",
        version = "v0.0.0-20191204190536-9bdfabe68543",
    )
    go_repository(
        name = "org_uber_go_goleak",
        importpath = "go.uber.org/goleak",
        sum = "h1:2K3zAYmnTNqV73imy9J1T3WC+gmCePx2hEGkimedGto=",
        version = "v1.3.0",
    )
    go_repository(
        name = "tools_gotest_v3",
        importpath = "gotest.tools/v3",
        sum = "h1:7koQfIKdy+I8UTetycgUqXWSDwpgv193Ka+qRsmBY8Q=",
        version = "v3.5.2",
    )

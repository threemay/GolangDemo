load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "nftprocess",
    srcs = [
        "nftmint.go",
        "util.go",
        "wire_gen.go",
        "workflow.go",
    ],
    importpath = "Mint/nftprocess",
    visibility = ["//:__subpackages__"],
    deps = [
        "//internal/pkg/configutil",
        "//internal/pkg/cryptoutil",
        "//internal/pkg/databaseutil",
        "//internal/pkg/domainutil",
        "//internal/pkg/nftutil",
        "//internal/pkg/service",
        "//internal/pkg/service/http",
        "//internal/pkg/service/objectstore",
        "//internal/pkg/walletutil",
        "@com_github_aws_aws_sdk_go_v2_service_secretsmanager//:secretsmanager",
        "@com_github_aws_aws_sdk_go_v2_service_ssm//:ssm",
        "@com_github_ethereum_go_ethereum//common",
        "@com_github_google_uuid//:uuid",
        "@com_github_google_wire//:wire",
        "@com_github_shopspring_decimal//:decimal",
        "@com_github_sirupsen_logrus//:logrus",
        "@org_golang_google_protobuf//types/known/fieldmaskpb",
        "@org_golang_google_protobuf//types/known/timestamppb",
    ],
)

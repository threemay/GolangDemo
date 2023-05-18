load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "Mint_lib",
    srcs = ["mint.go"],
    importpath = "nftprocess/Mint",
    visibility = ["//visibility:private"],
    deps = [
        "//internal/app/workflow/nftprocess",
        "//internal/pkg/domainutil",
        "@com_github_aws_aws_lambda_go//lambda",
    ],
)

go_binary(
    name = "Mint",
    embed = [":Mint_lib"],
    visibility = ["//visibility:public"],
)

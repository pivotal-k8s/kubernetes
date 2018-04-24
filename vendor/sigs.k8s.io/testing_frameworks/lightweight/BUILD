load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "apiserver.go",
        "control_plane.go",
        "doc.go",
        "etcd.go",
        "kubectl.go",
    ],
    importpath = "sigs.k8s.io/testing_frameworks/lightweight",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/sigs.k8s.io/testing_frameworks/cluster:go_default_library",
        "//vendor/sigs.k8s.io/testing_frameworks/lightweight/internal:go_default_library",
    ],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [
        ":package-srcs",
        "//vendor/sigs.k8s.io/testing_frameworks/lightweight/internal:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)

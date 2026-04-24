const branch = process.env.GITHUB_REF.split("/").pop();

const config = {
  branches: ["master", { name: "develop", prerelease: true }],
  plugins: [
    [
      "@semantic-release/commit-analyzer",
      {
        preset: "conventionalcommits",
      },
    ],
    [
      "@semantic-release/release-notes-generator",
      {
        preset: "conventionalcommits",
        presetConfig: {
          types: [
            {
              type: "feat",
              section: "Features",
            },
            {
              type: "fix",
              section: "Bug Fixes",
            },
            {
              type: "refactor",
              section: "Refactoring",
            },
            {
              type: "test",
              section: "Tests",
              hidden: true,
            },
            {
              type: "ci",
              section: "CI",
              hidden: true,
            },
            {
              type: "docs",
              section: "Documentation",
              hidden: true,
            },
            {
              type: "chore",
              section: "Chores",
              hidden: true,
            },
          ],
        },
      },
    ],
  ],
};

if (branch === "master") {
  config.plugins.push("@semantic-release/changelog");
}

config.plugins.push(
  [
    "@semantic-release/exec",
    {
      prepareCmd: ".github/workflows/update-build.sh ${nextRelease.version}",
    },
  ],
  [
    "@semantic-release/git",
    {
      assets: ["CHANGELOG.md", "VERSION.md", "k8s"],
    },
  ]
);

module.exports = config;

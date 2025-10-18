plugin "aws" {
  enabled = true
  version = "0.43.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}

plugin "terraform" {
  enabled = true
  version = "0.10.0"
  source  = "github.com/terraform-linters/tflint-ruleset-terraform"
}

config {
  module = true
  force  = false
}

rule "terraform_naming_convention" {
  enabled = true

  variable {
    format = "snake_case"
  }

  output {
    format = "snake_case"
  }

  resource {
    format = "snake_case"
  }
}

rule "terraform_documented_variables" {
  enabled = true
}

rule "terraform_documented_outputs" {
  enabled = true
}

rule "terraform_required_version" {
  enabled = true
}

rule "terraform_required_providers" {
  enabled = true
}

rule "terraform_standard_module_structure" {
  enabled = true
}

rule "terraform_unused_declarations" {
  enabled = true
}

rule "terraform_deprecated_interpolation" {
  enabled = true
}

rule "terraform_comment_syntax" {
  enabled = true
}

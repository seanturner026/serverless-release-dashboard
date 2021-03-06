variable "tags" {
  type        = map(string)
  description = "Map of tags to be applied to resources."
}

variable "admin_user_email" {
  type        = string
  description = <<-DESC
  Controls the creation of an admin user that is required to initially gain access to the
  dashboard.

  If access to the dashboard is completely lost, do the following
  • `var.enable_delete_admin_user = true`
  • `terraform apply`
  • `var.enable_delete_admin_user = false`
  • `terraform apply`

  If the initial admin user should no longer be able to access the dashboard, revoke access by
  setting `var.enable_delete_admin_user = true` and running `terraform apply`
  DESC
  default     = ""
}

variable "github_token" {
  type        = string
  description = "Token for Github."
}

variable "gitlab_token" {
  type        = string
  description = "Token for Gitlab."
}

variable "slack_webhook_url" {
  type        = string
  description = "URL to send slack message payloads to."
}

# Group Membership Request

Create a ServiceNow catalog request for adding a user to a group:

```sh
bam ticket group-membership <group-name> <user-name>
```

The command prints the ServiceNow request number after a successful submission. Before use, edit the ServiceNow settings in `cmd/ticket/group_membership.go`: tenant, API username and password, group-membership catalog item sys_id, and the catalog variable names for the group and user. These settings are code-level values and are not exposed as CLI options.

The ServiceNow user must be permitted to order the configured catalog item through the Service Catalog REST API. The catalog item should provide variables matching the configured group and user variable names.
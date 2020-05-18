package digitalocean

import (
	"context"
	"fmt"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDigitalOceanDatabaseReplica_Basic(t *testing.T) {
	t.Parallel()

	var databaseReplica godo.DatabaseReplica
	databaseName := randomTestName()
	databaseReplicaName := randomTestName()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDigitalOceanDatabaseReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccCheckDigitalOceanDatabaseReplicaConfigBasic, databaseName, databaseReplicaName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDigitalOceanDatabaseReplicaExists("digitalocean_database_replica.read-01", &databaseReplica),
					testAccCheckDigitalOceanDatabaseReplicaAttributes(&databaseReplica, databaseReplicaName),
					resource.TestCheckResourceAttr(
						"digitalocean_database_replica.read-01", "size", "db-s-2vcpu-4gb"),
					resource.TestCheckResourceAttr(
						"digitalocean_database_replica.read-01", "region", "nyc3"),
					resource.TestCheckResourceAttr(
						"digitalocean_database_replica.read-01", "name", databaseReplicaName),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_replica.read-01", "host"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_replica.read-01", "private_host"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_replica.read-01", "port"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_replica.read-01", "user"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_replica.read-01", "uri"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_replica.read-01", "private_uri"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_replica.read-01", "password"),
					resource.TestCheckResourceAttr(
						"digitalocean_database_replica.read-01", "tags.#", "1"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_replica.read-01", "private_network_uuid"),
				),
			},
		},
	})
}

func TestAccDigitalOceanDatabaseReplica_WithVPC(t *testing.T) {
	t.Parallel()

	var database godo.Database
	vpcName := randomTestName()
	databaseName := randomTestName()
	databaseReplicaName := randomTestName()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDigitalOceanDatabaseClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccCheckDigitalOceanDatabaseClusterConfigWithVPC, vpcName, databaseName) +
					fmt.Sprintf(testAccCheckDigitalOceanDatabaseReplicaConfigWithVPC, databaseReplicaName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDigitalOceanDatabaseClusterExists("digitalocean_database_cluster.foobar", &database),
					testAccCheckDigitalOceanDatabaseClusterAttributes(&database, databaseName),
					resource.TestCheckResourceAttrPair(
						"digitalocean_database_cluster.foobar", "private_network_uuid", "digitalocean_vpc.foobar", "id"),
					resource.TestCheckResourceAttrPair(
						"digitalocean_database_replica.read-01", "private_network_uuid", "digitalocean_vpc.foobar", "id"),
				),
			},
		},
	})
}

func testAccCheckDigitalOceanDatabaseReplicaDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*CombinedConfig).godoClient()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "digitalocean_database_replica" {
			continue
		}
		clusterId := rs.Primary.Attributes["cluster_id"]
		name := rs.Primary.Attributes["name"]
		// Try to find the database replica
		_, _, err := client.Databases.GetReplica(context.Background(), clusterId, name)

		if err == nil {
			return fmt.Errorf("DatabaseReplica still exists")
		}
	}

	return nil
}

func testAccCheckDigitalOceanDatabaseReplicaExists(n string, database *godo.DatabaseReplica) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DatabaseReplica cluster ID is set")
		}

		client := testAccProvider.Meta().(*CombinedConfig).godoClient()
		clusterId := rs.Primary.Attributes["cluster_id"]
		name := rs.Primary.Attributes["name"]

		foundDatabaseReplica, _, err := client.Databases.GetReplica(context.Background(), clusterId, name)

		if err != nil {
			return err
		}

		if foundDatabaseReplica.Name != name {
			return fmt.Errorf("DatabaseReplica not found")
		}

		*database = *foundDatabaseReplica

		return nil
	}
}

func testAccCheckDigitalOceanDatabaseReplicaAttributes(databaseReplica *godo.DatabaseReplica, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if databaseReplica.Name != name {
			return fmt.Errorf("Bad name: %s", databaseReplica.Name)
		}

		return nil
	}
}

const testAccCheckDigitalOceanDatabaseReplicaConfigBasic = `
resource "digitalocean_database_cluster" "foobar" {
	name       = "%s"
	engine     = "pg"
	version    = "11"
	size       = "db-s-1vcpu-1gb"
	region     = "nyc1"
	node_count = 1

	maintenance_window {
        day  = "friday"
        hour = "13:00:00"
	}
}

resource "digitalocean_database_replica" "read-01" {
  cluster_id = "${digitalocean_database_cluster.foobar.id}"
  name       = "%s"
  region     = "nyc3"
  size       = "db-s-2vcpu-4gb"
  tags       =	["staging"]
}`

const testAccCheckDigitalOceanDatabaseReplicaConfigWithVPC = `

resource "digitalocean_database_replica" "read-01" {
  cluster_id = digitalocean_database_cluster.foobar.id
  name       = "%s"
  region     = "nyc1"
  size       = "db-s-2vcpu-4gb"
  tags       =	["staging"]
  private_network_uuid = digitalocean_vpc.foobar.id
}`

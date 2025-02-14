<template>
	<v-card
		id="QueryTables"
		class="ma-0 pa-0 bg fill-height pane shadow-pane"
	>
		<v-card-title class="ma-0 py-0 py-sm-2 pl-3 pr-0 d-flex justify-start align-center">
			<v-icon
				:class="{
					'gray--text text--lighten-1': !$vuetify.theme.dark,
					'gray--text text--lighten-4': $vuetify.theme.dark,
				}"
				:size="24"
			>
				{{ mdiDatabaseOutline }}
			</v-icon>
			<UiActionDatabaseSelect
				v-if="currentDB"
				class="ma-0 py-0 pt-1 pl-2"
				dense
				:initial-value="currentDB"
				:prepend="$t('query.tables.title')"
				@update="onUpdateDatabase"
			/>
		</v-card-title>
		<v-card-text
			class="ma-0 pa-0 bg-secondary custom-scrollbar"
		>
			<v-treeview
				v-if="itemsLoaded && !pending"
				class="pt-3"
				:open.sync="open"
				:items="items"
				item-key="id"
				open-on-click
				hoverable
				:multiple-active="false"
			>
				<template
					slot="label"
					slot-scope="props"
				>
					<span
						class="ma-0 pa-0 d-flex justify-space-start align-center fill-width"
					>
						<v-tooltip
							content-class="ma-0 py-2 px-4 bg primary-outlined arrow-bottom-center"
							top
							:open-delay="300"
							:nudge-top="6"
						>
							<template #activator="{ on, attrs }">
								<v-icon
									v-if="props.item.type !== 'node'"
									:class="`ma-0 mr-2 pa-0 ${ getColor(props.item.type) }`"
									small
									v-bind="attrs"
									v-on="on"
								>
									{{ getIcon(props.item.type) }}
								</v-icon>
							</template>
							<span>
								{{ $t(`query.tables.${ props.item.type }`) }}
							</span>
						</v-tooltip>
						<span
							class="h-32 sql-column body-2 d-flex align-center fill-width"
							:class="{
								'gray--text text--darken-3': !$vuetify.theme.dark,
								'gray--text text--lighten-5': $vuetify.theme.dark,
							}"
						>
							<span class="h-32 d-flex align-center">
								{{ props.item.label }}
							</span>
						</span>
						<v-spacer />
						<QueryTablesInfo
							:item="props.item"
							@update="onQueryUpdate"
						/>
					</span>
				</template>
			</v-treeview>
			<QueryTablesEmpty
				v-else-if="!pending"
				:allowed="showAdd"
				@submit="onAddDemoData"
			/>
			<QueryTablesSkeleton
				v-else
				class="ma-0 pa-4"
			/>
		</v-card-text>
	</v-card>
</template>

<script>
import { mapActions, mapGetters } from 'vuex';
import {
	AUTH_MODULE,
	USER_PERMISSION,
} from '@/store/auth/constants';
import {
	DATABASE_MODULE,
	TABLE_LIST,
} from '@/store/database/constants';
import {
	IMMUDB_MODULE,
	RUN_SQL_EXEC,
	STATE,
} from '@/store/immudb/constants';
import {
	mdiDatabaseOutline,
	mdiTable,
	mdiTableColumn,
} from '@mdi/js';

export default {
	name: 'QueryTables',
	props: {
		pending: { type: Boolean, default: false },
	},
	data () {
		return {
			mdiDatabaseOutline,
			mdiTable,
			mdiTableColumn,
			open: [],
			items: [],
		};
	},
	computed: {
		...mapGetters(AUTH_MODULE, {
			userPermission: USER_PERMISSION,
		}),
		...mapGetters(DATABASE_MODULE, {
			tables: TABLE_LIST,
		}),
		...mapGetters(IMMUDB_MODULE, {
			state: STATE,
		}),
		showAdd () {
			try {
				const { permission } = this.userPermission(this.currentDB) ||
					{ permission: 0 };
				return permission > 1;
			}
			catch (err) {
				console.error(err);
				return false;
			}
		},
		currentDB () {
			if (this.state) {
				const { db } = this.state;
				return db;
			}
			return '';
		},
		itemsLoaded () {
			if (this.items) {
				if (this.items.length > 0) {
					return true;
				}
				else if (this.items[0]) {
					const { children } = this.items[0];
					return children && children.length > 0;
				}
			}
			return false;
		},
	},
	watch: {
		tables: {
			deep: true,
			immediate: true,
			handler (newVal) {
				if (newVal) {
					const children = this.parseTables(newVal);
					if (children) {
						this.items = children;
					}
				}
			},
		},
	},
	methods: {
		...mapActions(IMMUDB_MODULE, {
			runSqlExec: RUN_SQL_EXEC,
		}),
		parseTables (data) {
			if (data) {
				const _data = JSON.parse(JSON.stringify(data));
				if (_data && _data.length) {
					return _data
							.slice()
							.sort((a, b) => a.label < b.label ? -1 : 1)
							.sort((a, b) => a.primaryKey && !b.primaryKey ? -1 : 1)
							.sort((a, b) => a.foreignKey && !b.foreignKey ? -1 : 1)
							.map((_) => {
								const t = _;
								if (_.children && _.children.length) {
									t.children = this.parseTables(t.children);
									return t;
								}
								return t;
							});
				}
			}
			return [];
		},
		getIcon (data) {
			if (data === 'table' || data === 'tab') {
				return mdiTable;
			}
			else if (data === 'column' || data === 'col') {
				return mdiTableColumn;
			}
			else {
				return mdiTable;
			}
		},
		getColor (data) {
			if (data === 'table' || data === 'tab') {
				return 'primary--text';
			}
			else if (data === 'column' || data === 'col') {
				return `primary--text ${ this.$vuetify.theme.dark ? 'text--lighten-3' : 'text--lighten-1' }`;
			}
			else {
				return 'grey';
			}
		},
		onQueryUpdate ({ value }) {
			this.$eventbus && this.$eventbus
					.$emit('EVENT_BUS==>updateQuery', {
						uid: this._uid,
						value,
					});
			setTimeout(() => {
				this.$eventbus && this.$eventbus
						.$emit('EVENT_BUS==>focusQuery');
			}, 30);
		},
		async onAddDemoData () {
			try {
				this.$emit('update:start');

				/* eslint-disable quotes */
				// create and fill customers table with random data
				await this.runSqlExec("CREATE TABLE customers (id INTEGER, customer_name VARCHAR, email VARCHAR, address VARCHAR, city VARCHAR, ip VARCHAR, country VARCHAR, age INTEGER, active BOOLEAN, PRIMARY KEY id);");
				await this.runSqlExec("UPSERT INTO customers (id, customer_name, email, address, city, ip, country, age, active) values (1, 'Isidro Behnen', 'ibehnen0@mail.ru', 'ibehnen0@chronoengine.com', 'Arvika', '2.124.67.107', 'SE', 24, true);");
				await this.runSqlExec("UPSERT INTO customers (id, customer_name, email, address, city, ip, country, age, active) values (2, 'Claudianus Boldt', 'cboldt1@adobe.com', 'cboldt1@elpais.com', 'Kimhae', '125.89.31.130', 'KR', 30, true);");
				await this.runSqlExec("UPSERT INTO customers (id, customer_name, email, address, city, ip, country, age, active) values (3, 'Ginny Kelshaw', 'gkelshaw2@bbb.org', 'gkelshaw2@123-reg.co.uk', 'Nova Lima', '231.174.212.41', 'BR', 32, true);");
				await this.runSqlExec("UPSERT INTO customers (id, customer_name, email, address, city, ip, country, age, active) values (4, 'Paola Elverston', 'pelverston3@ox.ac.uk', 'pelverston3@cdc.gov', 'Morales', '114.130.229.206', 'GT', 33, true);");
				await this.runSqlExec("UPSERT INTO customers (id, customer_name, email, address, city, ip, country, age, active) values (1, 'Asa Antrum', 'aantrum0@apple.com', 'aantrum0@prweb.com', 'Selkirk', '193.5.161.248', 'CA', 46, true);");
				await this.runSqlExec("UPSERT INTO customers (id, customer_name, email, address, city, ip, country, age, active) values (2, 'Neila Rowena', 'nrowena1@google.ca', 'nrowena1@netscape.com', 'Liangdong', '65.163.125.159', 'CN', 39, true);");
				await this.runSqlExec("UPSERT INTO customers (id, customer_name, email, address, city, ip, country, age, active) values (3, 'Rex Massinger', 'rmassinger2@weibo.com', 'rmassinger2@jiathis.com', 'Jesús Menéndez', '165.100.211.115', 'CU', 28, false);");
				await this.runSqlExec("UPSERT INTO customers (id, customer_name, email, address, city, ip, country, age, active) values (4, 'Whit Shilton', 'wshilton3@amazonaws.com', 'wshilton3@cdbaby.com', 'Bobowo', '167.156.190.237', 'PL', 52, false);");

				// create products table with random data
				await this.runSqlExec("CREATE TABLE products (id INTEGER, product VARCHAR, price VARCHAR, PRIMARY KEY id);");
				await this.runSqlExec("UPSERT INTO products (id, product, price) values (1, 'Juice - V8, Tomato', '$4.04');");
				await this.runSqlExec("UPSERT INTO products (id, product, price) values (2, 'Grapes - Red', '$5.03');");
				await this.runSqlExec("UPSERT INTO products (id, product, price) values (3, 'Compound - Mocha', '$6.75');");
				await this.runSqlExec("UPSERT INTO products (id, product, price) values (4, 'Pear - Asian', '$1.10');");
				await this.runSqlExec("UPSERT INTO products (id, product, price) values (5, 'Wine - Mas Chicet Rose, Vintage', '$6.75');");
				await this.runSqlExec("UPSERT INTO products (id, product, price) values (6, 'Lady Fingers', '$5.32');");
				await this.runSqlExec("UPSERT INTO products (id, product, price) values (7, 'Shrimp - 31/40', '$9.04');");
				await this.runSqlExec("UPSERT INTO products (id, product, price) values (8, 'Pork - Bones', '$2.65');");
				await this.runSqlExec("UPSERT INTO products (id, product, price) values (9, 'Table Cloth 62x114 White', '$0.73');");
				await this.runSqlExec("UPSERT INTO products (id, product, price) values (10, 'Beef - Ground, Extra Lean, Fresh', '$7.29');");

				// create orders table with random data
				await this.runSqlExec("CREATE TABLE orders (id INTEGER, customerID INTEGER, productID INTEGER, PRIMARY KEY id);");
				await this.runSqlExec("UPSERT INTO orders (id, customerID, productID) VALUES (1, 1, 1);");
				await this.runSqlExec("UPSERT INTO orders (id, customerID, productID) VALUES (2, 5, 2);");
				await this.runSqlExec("UPSERT INTO orders (id, customerID, productID) VALUES (3, 4, 3);");
				await this.runSqlExec("UPSERT INTO orders (id, customerID, productID) VALUES (4, 3, 4);");
				await this.runSqlExec("UPSERT INTO orders (id, customerID, productID) VALUES (5, 2, 5);");
				/* eslint-enable quotes */

				this.$emit('update:end');
			}
			catch (err) {
				console.error(err);
			}
		},
		onUpdateDatabase (data) {
			this.$emit('update:use', data);
		},
	},
};
</script>

<style lang="scss">
#QueryTables {
	.v-treeview {
		padding-left: 2px !important;

		.sql-column {
			white-space: nowrap !important;
			overflow: hidden !important;
			text-overflow: ellipsis !important;
		}

		&.theme-- {
			&light {
				.v-treeview-node__toggle {
					color: var(--v-primary-base) !important;
				}

				.v-treeview-node__label {
					a {
						color: #191919 !important;

						&.nuxt-link-exact-active {
							color: #191919 !important;
						}

						&::before {
							background-color: var(--v-primary-base);
						}

						&::after {
							background-color: rgba(25, 119, 210, 0.15) !important;
						}
					}
				}

				.v-treeview-node__children {
					&::before {
						background-color: rgba(0, 0, 0, 0.25);
					}
				}
			}

			&dark {
				.v-treeview-node__toggle {
					color: var(--v-primary-base) !important;
				}

				.v-treeview-node__label {
					a {
						color: #e6e6e6 !important;

						&.nuxt-link-exact-active {
							color: #e6e6e6 !important;
						}

						&::before {
							background-color: var(--v-primary-base);
						}

						&::after {
							background-color: rgba(25, 119, 210, 0.15) !important;
						}
					}
				}

				.v-treeview-node__children {
					&::before {
						background-color: rgba(255, 255, 255, 0.25);
					}
				}
			}
		}

		.v-treeview-node__root {
			min-height: $spacer-8;

			.v-treeview-node__level {
				display: none !important;
			}

			.table-actions {
				.tags,
				.add-to-query {
					transition: opacity 0.11s ease-in-out;
				}

				.tags {
					display: flex !important;
					opacity: 1;
				}

				.add-to-query {
					display: none !important;
					opacity: 0;
				}
			}

			&:hover {
				// .tags {
				// 	display: none !important;
				// 	opacity: 0;
				// }

				.add-to-query {
					display: flex !important;
					opacity: 1;
				}
			}
		}

		.v-treeview-node__children {
			position: relative;
			padding-left: $spacer-5;

			.v-treeview-node--leaf {
				padding-left: $spacer-1;
			}

			&::before {
				content: '';
				position: absolute;
				top: 0;
				bottom: 0;
				left: calc(#{$spacer-5} - 1px);
				width: 2px;
			}
		}

		.v-treeview-node__content {
			.v-treeview-node__label {
				position: relative;
				display: flex;
				height: 100%;
				width: 100%;
				padding-right: $spacer-1;
				font-size: 0.8rem;

				a {
					display: flex;
					flex-grow: 1;
					justify-content: flex-start;
					align-items: center;
					min-height: $spacer-8;
					padding: 0 $spacer-2 0 0 !important;

					&.nuxt-link-exact-active {
						&::before {
							content: '';
							position: absolute;
							top: 0;
							bottom: 0;
							left: -1px !important;
							width: 2px;
						}

						&::after {
							content: '';
							position: absolute;
							top: 0;
							right: 0;
							bottom: 0;
							left: 0;
						}
					}
				}
			}
		}
	}
}
</style>

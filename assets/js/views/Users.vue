<template>
	<div class="root safe-area-inset">
		<div class="container px-4">
			<TopHeader :title="$t('users.title')" :notifications="notifications" />
			<div class="wrapper pb-5">
				<h2 class="my-4 mt-5">{{ $t("users.section.users") }}</h2>
				<div class="p-0 config-list">
					<UserCard
						v-for="user in users"
						:key="user.id"
						:user="user"
						@edit="openModal(user)"
					/>
					<NewDeviceButton
						:title="$t('users.addUser')"
						data-testid="add-user"
						@click="openModal()"
					/>
				</div>
			</div>
		</div>
		<UserModal ref="userModal" :vehicles="vehicleNames" :loadpoints="loadpointNames" @changed="loadUsers" />
	</div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import Header from "../components/Top/Header.vue";
import NewDeviceButton from "../components/Config/NewDeviceButton.vue";
import UserCard from "../components/Config/UserCard.vue";
import UserModal from "../components/Config/UserModal.vue";
import api from "../api";
import type { Notification, User } from "@/types/evcc";

interface ConfigVehicle {
	name: string;
	config?: { title?: string };
}

export default defineComponent({
	name: "UsersView",
	components: {
		TopHeader: Header,
		NewDeviceButton,
		UserCard,
		UserModal,
	},
	props: {
		notifications: { type: Array as PropType<Notification[]>, default: () => [] },
	},
	data(): { users: User[]; vehicleNames: string[]; loadpointNames: string[] } {
		return {
			users: [],
			vehicleNames: [],
			loadpointNames: [],
		};
	},
	async mounted() {
		await Promise.all([this.loadUsers(), this.loadVehicles(), this.loadLoadpoints()]);
	},
	methods: {
		async loadUsers() {
			try {
				const res = await api.get("/users");
				this.users = res.data ?? [];
			} catch (e) {
				console.error("Failed to load users", e);
			}
		},
		async loadVehicles() {
			try {
				const res = await api.get("/config/devices/vehicle");
				const vehicles: ConfigVehicle[] = res.data ?? [];
				this.vehicleNames = vehicles.map((v) => v.config?.title || v.name);
			} catch {
				this.vehicleNames = [];
			}
		},
		async loadLoadpoints() {
			try {
				const res = await api.get("/config/loadpoints");
				const loadpoints: { name?: string; title?: string }[] = res.data ?? [];
				this.loadpointNames = loadpoints.map((lp) => lp.title || lp.name || "");
			} catch {
				this.loadpointNames = [];
			}
		},
		openModal(user?: User) {
			const defaultRole = !user && this.users.length === 0 ? "admin" : undefined;
			(this.$refs["userModal"] as unknown as InstanceType<typeof UserModal>).open(
				user,
				defaultRole
			);
		},
	},
});
</script>

<style scoped>
.config-list {
	display: grid;
	grid-template-columns: repeat(auto-fill, minmax(min(300px, 100%), 1fr));
	gap: 1rem;
}
</style>

<template>
	<GenericModal
		id="userModal"
		:title="isEdit ? $t('users.modal.titleEdit') : $t('users.modal.titleAdd')"
		data-testid="user-modal"
		@open="reset"
		@closed="closed"
	>
		<form @submit.prevent="save">
			<div class="mb-3">
				<label class="form-label" for="userModalUsername">
					{{ $t("users.modal.username") }}
				</label>
				<input
					id="userModalUsername"
					v-model="username"
					type="text"
					class="form-control"
					autocomplete="username"
					:disabled="isEdit"
					required
				/>
			</div>

			<div class="mb-3">
				<label class="form-label" for="userModalRole">
					{{ $t("users.modal.role") }}
				</label>
				<select id="userModalRole" v-model="role" class="form-select">
					<option v-for="r in roles" :key="r.value" :value="r.value">
						{{ r.label }}
					</option>
				</select>
			</div>

			<div v-if="!isEdit" class="mb-3">
				<label class="form-label" for="userModalPassword">
					{{ $t("users.modal.password") }}
				</label>
				<input
					id="userModalPassword"
					v-model="password"
					type="password"
					class="form-control"
					autocomplete="new-password"
					required
				/>
			</div>

			<div v-if="!isEdit" class="mb-3">
				<label class="form-label" for="userModalPasswordRepeat">
					{{ $t("users.modal.passwordRepeat") }}
				</label>
				<input
					id="userModalPasswordRepeat"
					v-model="passwordRepeat"
					type="password"
					class="form-control"
					autocomplete="new-password"
					required
				/>
			</div>

			<div v-if="vehicles.length > 0 || loadpoints.length > 0" class="mb-3 row g-3">
				<!-- vehicles column -->
				<div class="col-6">
					<label class="form-label">{{ $t("users.modal.vehicles") }}</label>

					<div class="d-flex flex-wrap gap-2 mb-2">
						<span
							v-for="v in selectedVehicles"
							:key="v"
							class="badge rounded-pill bg-primary d-flex align-items-center gap-1 px-2 py-2"
						>
							<shopicon-regular-car1 size="xs"></shopicon-regular-car1>
							{{ v }}
							<button
								type="button"
								class="btn-close btn-close-white ms-1"
								style="font-size: 0.6em"
								:aria-label="$t('users.modal.vehicleRemove')"
								@click="removeVehicle(v)"
							></button>
						</span>
						<span v-if="selectedVehicles.length === 0" class="text-muted small">
							{{ $t("users.modal.vehiclesNone") }}
						</span>
					</div>

					<div v-if="availableVehicles.length > 0" class="add-list">
						<button
							v-for="v in availableVehicles"
							:key="v"
							type="button"
							class="add-item d-flex align-items-center gap-2 w-100 text-start border-0 bg-transparent py-1 px-2 rounded"
							@click="addVehicle(v)"
						>
							<shopicon-regular-plus size="xs" class="text-success flex-shrink-0"></shopicon-regular-plus>
							<span>{{ v }}</span>
						</button>
					</div>
				</div>

				<!-- loadpoints column -->
				<div class="col-6">
					<label class="form-label">{{ $t("users.modal.loadpoints") }}</label>

					<div class="d-flex flex-wrap gap-2 mb-2">
						<span
							v-for="lp in selectedLoadpoints"
							:key="lp"
							class="badge rounded-pill bg-success d-flex align-items-center gap-1 px-2 py-2"
						>
							<shopicon-regular-cablecharge size="xs"></shopicon-regular-cablecharge>
							{{ lp }}
							<button
								type="button"
								class="btn-close btn-close-white ms-1"
								style="font-size: 0.6em"
								:aria-label="$t('users.modal.loadpointRemove')"
								@click="removeLoadpoint(lp)"
							></button>
						</span>
						<span v-if="selectedLoadpoints.length === 0" class="text-muted small">
							{{ $t("users.modal.loadpointsNone") }}
						</span>
					</div>

					<div v-if="availableLoadpoints.length > 0" class="add-list">
						<button
							v-for="lp in availableLoadpoints"
							:key="lp"
							type="button"
							class="add-item d-flex align-items-center gap-2 w-100 text-start border-0 bg-transparent py-1 px-2 rounded"
							@click="addLoadpoint(lp)"
						>
							<shopicon-regular-plus size="xs" class="text-success flex-shrink-0"></shopicon-regular-plus>
							<span>{{ lp }}</span>
						</button>
					</div>
				</div>
			</div>

			<div v-if="error" class="text-danger mb-3">{{ error }}</div>

			<div class="d-flex justify-content-between">
				<button
					v-if="isEdit"
					type="button"
					class="btn btn-danger"
					:disabled="saving"
					@click="deleteUser"
				>
					{{ $t("users.modal.delete") }}
				</button>
				<div v-else></div>
				<button type="submit" class="btn btn-primary" :disabled="saving">
					{{ $t("users.modal.save") }}
				</button>
			</div>
		</form>
	</GenericModal>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import GenericModal from "../Helper/GenericModal.vue";
import Modal from "bootstrap/js/dist/modal";
import api from "../../api";
import type { User } from "@/types/evcc";
import "@h2d2/shopicons/es/regular/car1";
import "@h2d2/shopicons/es/regular/plus";
import "@h2d2/shopicons/es/regular/cablecharge";

type UserData = Partial<User>;

export default defineComponent({
	name: "UserModal",
	components: { GenericModal },
	props: {
		vehicles: { type: Array as PropType<string[]>, default: () => [] },
		loadpoints: { type: Array as PropType<string[]>, default: () => [] },
	},
	emits: ["changed"],
	data() {
		return {
			editUser: null as UserData | null,
			defaultRole: "viewer" as string,
			username: "",
			role: "viewer",
			password: "",
			passwordRepeat: "",
			selectedVehicles: [] as string[],
			selectedLoadpoints: [] as string[],
			error: "",
			saving: false,
		};
	},
	computed: {
		isEdit() {
			return !!this.editUser;
		},
		availableVehicles(): string[] {
			return this.vehicles.filter((v) => !this.selectedVehicles.includes(v));
		},
		availableLoadpoints(): string[] {
			return this.loadpoints.filter((lp) => !this.selectedLoadpoints.includes(lp));
		},
		roles(): { value: string; label: string }[] {
			return [
				{ value: "viewer", label: this.$t("users.role.viewer") },
				{ value: "user", label: this.$t("users.role.user") },
				{ value: "maintainer", label: this.$t("users.role.maintainer") },
				{ value: "admin", label: this.$t("users.role.admin") },
			];
		},
	},
	methods: {
		open(user?: UserData, defaultRole?: string) {
			this.editUser = user ?? null;
			this.defaultRole = defaultRole ?? "viewer";
			const el = document.getElementById("userModal");
			if (el) Modal.getOrCreateInstance(el).show();
		},
		close() {
			const el = document.getElementById("userModal");
			if (el) Modal.getOrCreateInstance(el).hide();
		},
		reset() {
			if (this.editUser) {
				this.username = this.editUser.username ?? "";
				this.role = this.editUser.role ?? "viewer";
				this.selectedVehicles = this.editUser.vehicles ? [...this.editUser.vehicles] : [];
				this.selectedLoadpoints = this.editUser.loadpoints ? [...this.editUser.loadpoints] : [];
			} else {
				this.username = "";
				this.role = this.defaultRole;
				this.selectedVehicles = [];
				this.selectedLoadpoints = [];
			}
			this.password = "";
			this.passwordRepeat = "";
			this.error = "";
			this.saving = false;
		},
		closed() {
			this.editUser = null;
		},
		async save() {
			this.error = "";
			if (!this.isEdit) {
				if (!this.password) {
					this.error = this.$t("users.modal.errorPasswordEmpty");
					return;
				}
				if (this.password !== this.passwordRepeat) {
					this.error = this.$t("users.modal.errorPasswordMismatch");
					return;
				}
			}
			this.saving = true;
			try {
				if (this.isEdit && this.editUser?.id) {
					await api.patch(`/users/${this.editUser.id}`, { role: this.role });
					await Promise.all([
						this.syncVehicles(this.editUser.id),
						this.syncLoadpoints(this.editUser.id),
					]);
				} else {
					const res = await api.post("/users", {
						username: this.username,
						password: this.password,
						role: this.role,
					});
					if (res.data?.id) {
						await Promise.all([
							this.syncVehicles(res.data.id),
							this.syncLoadpoints(res.data.id),
						]);
					}
				}
				this.$emit("changed");
				this.close();
			} catch (e: any) {
				this.error = e?.response?.data || e?.message || String(e);
			} finally {
				this.saving = false;
			}
		},
		addVehicle(v: string) {
			if (!this.selectedVehicles.includes(v)) {
				this.selectedVehicles.push(v);
			}
		},
		removeVehicle(v: string) {
			this.selectedVehicles = this.selectedVehicles.filter((x) => x !== v);
		},
		addLoadpoint(lp: string) {
			if (!this.selectedLoadpoints.includes(lp)) {
				this.selectedLoadpoints.push(lp);
			}
		},
		removeLoadpoint(lp: string) {
			this.selectedLoadpoints = this.selectedLoadpoints.filter((x) => x !== lp);
		},
		async syncVehicles(userId: number) {
			const current = this.editUser?.vehicles ?? [];
			const add = this.selectedVehicles.filter((v) => !current.includes(v));
			const remove = current.filter((v) => !this.selectedVehicles.includes(v));
			await Promise.all([
				...add.map((v) => api.post(`/users/${userId}/vehicles`, { vehicleName: v })),
				...remove.map((v) => api.delete(`/users/${userId}/vehicles/${encodeURIComponent(v)}`)),
			]);
		},
		async syncLoadpoints(userId: number) {
			const current = this.editUser?.loadpoints ?? [];
			const add = this.selectedLoadpoints.filter((lp) => !current.includes(lp));
			const remove = current.filter((lp) => !this.selectedLoadpoints.includes(lp));
			await Promise.all([
				...add.map((lp) =>
					api.post(`/users/${userId}/loadpoints`, { loadpointName: lp })
				),
				...remove.map((lp) =>
					api.delete(`/users/${userId}/loadpoints/${encodeURIComponent(lp)}`)
				),
			]);
		},
		async deleteUser() {
			if (!this.editUser?.id) return;
			this.saving = true;
			try {
				await api.delete(`/users/${this.editUser.id}`);
				this.$emit("changed");
				this.close();
			} catch (e: any) {
				this.error = e?.response?.data || e?.message || String(e);
			} finally {
				this.saving = false;
			}
		},
	},
});
</script>

<style scoped>
.add-list {
	border: 1px solid var(--bs-border-color);
	border-radius: var(--bs-border-radius);
	overflow: hidden;
}
.add-item {
	color: var(--bs-body-color);
	font-size: 0.9em;
}
.add-item:hover {
	background-color: var(--bs-secondary-bg) !important;
}
</style>

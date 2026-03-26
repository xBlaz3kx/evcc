<template>
	<DeviceCard :title="user.username" editable @edit="$emit('edit')">
		<template #icon>
			<shopicon-regular-account></shopicon-regular-account>
		</template>
		<template #tags>
			<div class="d-flex flex-wrap align-items-center gap-2">
				<span class="badge rounded-pill" :class="roleBadgeClass">
					{{ $t(`users.role.${user.role}`) }}
				</span>
				<span
					v-for="vehicle in user.vehicles"
					:key="`v_${vehicle}`"
					class="badge rounded-pill bg-primary d-flex align-items-center gap-1"
				>
					<shopicon-regular-car1 size="xs"></shopicon-regular-car1>
					{{ vehicle }}
				</span>
				<span
					v-for="lp in user.loadpoints"
					:key="`lp_${lp}`"
					class="badge rounded-pill bg-success d-flex align-items-center gap-1"
				>
					<shopicon-regular-cablecharge size="xs"></shopicon-regular-cablecharge>
					{{ lp }}
				</span>
			</div>
		</template>
	</DeviceCard>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import "@h2d2/shopicons/es/regular/account";
import "@h2d2/shopicons/es/regular/car1";
import "@h2d2/shopicons/es/regular/cablecharge";
import DeviceCard from "./DeviceCard.vue";
import type { User } from "@/types/evcc";

export default defineComponent({
	name: "UserCard",
	components: { DeviceCard },
	props: {
		user: { type: Object as PropType<User>, required: true },
	},
	emits: ["edit"],
	computed: {
		roleBadgeClass(): string {
			const map: Record<string, string> = {
				admin: "bg-danger",
				maintainer: "bg-warning text-dark",
				user: "bg-primary",
				viewer: "bg-secondary",
			};
			return map[this.user.role] ?? "bg-secondary";
		},
	},
});
</script>

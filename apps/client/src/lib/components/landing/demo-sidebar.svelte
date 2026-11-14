<script lang="ts">
	import { Bar } from '@facile/mise-en-scene';
	import { Icon } from '@facile/muse';
</script>

<div class="vd-nav">
	<span class="vd-brand"><Icon icon="solar:panorama-bold-duotone" size={18} />Vision</span>
	<span class="vd-nav-row vd-active"><Bar w="62%" /></span>
	<span class="vd-nav-row"><Bar w="48%" /></span>
	<span class="vd-nav-row"><Bar w="56%" /></span>
	<span class="vd-nav-row"><Bar w="40%" /></span>
</div>

<style>
	.vd-nav {
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding: 16px 12px;
	}
	.vd-brand {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 0 8px 12px;
		font-family: var(--font-fc-title);
		font-size: 14px;
		font-weight: 700;
	}
	.vd-nav-row {
		display: flex;
		align-items: center;
		height: 32px;
		padding: 0 10px;
		border-radius: 8px;
	}
	.vd-active {
		background: var(--mes-w-body);
	}
</style>

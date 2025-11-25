<script lang="ts">
	export let status: string;

	$: badgeClass = {
		active: 'badge-active',
		fixed: 'badge-fixed',
		ignored: 'badge-ignored',
		accepted: 'badge-accepted'
	}[status] || 'badge-active';
</script>

<span class="badge {badgeClass}">
	{status}
</span>

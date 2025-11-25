<script lang="ts">
	export let severity: string;

	$: badgeClass = {
		Critical: 'badge-critical',
		High: 'badge-high',
		Medium: 'badge-medium',
		Low: 'badge-low'
	}[severity] || 'badge-low';
</script>

<span class="badge {badgeClass}">
	{severity}
</span>

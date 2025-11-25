<script lang="ts">
	import { onMount } from 'svelte';
	import { loadMetrics, metrics, metricsLoading } from '$lib/stores/metrics';
	import SeverityBadge from '$lib/components/SeverityBadge.svelte';

	onMount(() => {
		loadMetrics();
	});
</script>

<svelte:head>
	<title>Dashboard - Invulnerable</title>
</svelte:head>

<div class="space-y-6">
	<h1 class="text-3xl font-bold text-gray-900">Dashboard</h1>

	{#if $metricsLoading}
		<div class="text-center py-12">
			<p class="text-gray-500">Loading metrics...</p>
		</div>
	{:else if $metrics}
		<!-- Summary Cards -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
			<div class="card">
				<h3 class="text-sm font-medium text-gray-500">Total Images</h3>
				<p class="mt-2 text-3xl font-bold text-gray-900">{$metrics.total_images}</p>
			</div>

			<div class="card">
				<h3 class="text-sm font-medium text-gray-500">Total Scans</h3>
				<p class="mt-2 text-3xl font-bold text-gray-900">{$metrics.total_scans}</p>
			</div>

			<div class="card">
				<h3 class="text-sm font-medium text-gray-500">Active Vulnerabilities</h3>
				<p class="mt-2 text-3xl font-bold text-red-600">{$metrics.active_vulnerabilities}</p>
			</div>

			<div class="card">
				<h3 class="text-sm font-medium text-gray-500">Scans (24h)</h3>
				<p class="mt-2 text-3xl font-bold text-gray-900">{$metrics.recent_scans_24h}</p>
			</div>
		</div>

		<!-- Severity Breakdown -->
		<div class="card">
			<h2 class="text-xl font-bold text-gray-900 mb-4">Vulnerabilities by Severity</h2>
			<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
				<div class="flex items-center justify-between p-4 bg-red-50 rounded-lg">
					<div>
						<p class="text-sm font-medium text-red-800">Critical</p>
						<p class="text-2xl font-bold text-red-900">{$metrics.severity_counts.critical}</p>
					</div>
					<SeverityBadge severity="Critical" />
				</div>

				<div class="flex items-center justify-between p-4 bg-orange-50 rounded-lg">
					<div>
						<p class="text-sm font-medium text-orange-800">High</p>
						<p class="text-2xl font-bold text-orange-900">{$metrics.severity_counts.high}</p>
					</div>
					<SeverityBadge severity="High" />
				</div>

				<div class="flex items-center justify-between p-4 bg-yellow-50 rounded-lg">
					<div>
						<p class="text-sm font-medium text-yellow-800">Medium</p>
						<p class="text-2xl font-bold text-yellow-900">{$metrics.severity_counts.medium}</p>
					</div>
					<SeverityBadge severity="Medium" />
				</div>

				<div class="flex items-center justify-between p-4 bg-blue-50 rounded-lg">
					<div>
						<p class="text-sm font-medium text-blue-800">Low</p>
						<p class="text-2xl font-bold text-blue-900">{$metrics.severity_counts.low}</p>
					</div>
					<SeverityBadge severity="Low" />
				</div>
			</div>
		</div>

		<!-- Vulnerability Trend -->
		{#if $metrics.vulnerability_trend && $metrics.vulnerability_trend.length > 0}
			<div class="card">
				<h2 class="text-xl font-bold text-gray-900 mb-4">Vulnerability Trend (30 Days)</h2>
				<div class="overflow-x-auto">
					<table class="min-w-full divide-y divide-gray-200">
						<thead>
							<tr>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Date</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Count</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-200">
							{#each $metrics.vulnerability_trend.slice(0, 10) as point}
								<tr>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{point.date}</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{point.count}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		{/if}
	{:else}
		<div class="text-center py-12">
			<p class="text-gray-500">No metrics available</p>
		</div>
	{/if}
</div>

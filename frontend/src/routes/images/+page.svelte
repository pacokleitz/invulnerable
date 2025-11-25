<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import type { ImageWithStats } from '$lib/api/types';

	let images: ImageWithStats[] = [];
	let loading = true;
	let error: string | null = null;

	onMount(async () => {
		try {
			images = await api.images.list({ limit: 100 });
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load images';
		} finally {
			loading = false;
		}
	});

	function formatDate(date?: string): string {
		if (!date) return 'Never';
		return new Date(date).toLocaleString();
	}
</script>

<svelte:head>
	<title>Images - Invulnerable</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<h1 class="text-3xl font-bold text-gray-900">Images</h1>
	</div>

	{#if loading}
		<div class="text-center py-12">
			<p class="text-gray-500">Loading images...</p>
		</div>
	{:else if error}
		<div class="card bg-red-50">
			<p class="text-red-600">{error}</p>
		</div>
	{:else if images.length === 0}
		<div class="card text-center py-12">
			<p class="text-gray-500">No images found</p>
		</div>
	{:else}
		<div class="card overflow-hidden">
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-200">
					<thead class="bg-gray-50">
						<tr>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Image</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Scans</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Last Scan</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Critical</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">High</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Medium</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Low</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
						</tr>
					</thead>
					<tbody class="bg-white divide-y divide-gray-200">
						{#each images as image}
							<tr class="hover:bg-gray-50">
								<td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
									{image.registry}/{image.repository}:{image.tag}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{image.scan_count}</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
									{formatDate(image.last_scan_date)}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									<span class="text-red-600 font-semibold">{image.critical_count}</span>
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									<span class="text-orange-600 font-semibold">{image.high_count}</span>
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									<span class="text-yellow-600 font-semibold">{image.medium_count}</span>
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									<span class="text-blue-600 font-semibold">{image.low_count}</span>
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									<a href="/images/{image.id}" class="text-blue-600 hover:text-blue-800">View History</a>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>

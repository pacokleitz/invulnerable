<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';

	let sbom: any = null;
	let sbomJson: string = '';
	let loading = true;
	let error: string | null = null;
	let copied = false;

	$: scanId = parseInt($page.params.id);

	onMount(async () => {
		try {
			sbom = await api.scans.getSBOM(scanId);
			sbomJson = JSON.stringify(sbom, null, 2);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load SBOM';
		} finally {
			loading = false;
		}
	});

	function downloadSBOM() {
		const blob = new Blob([sbomJson], { type: 'application/json' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `scan-${scanId}-sbom.json`;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	}

	function copySBOM() {
		navigator.clipboard.writeText(sbomJson).then(() => {
			copied = true;
			setTimeout(() => {
				copied = false;
			}, 2000);
		});
	}
</script>

<svelte:head>
	<title>SBOM - Scan {scanId} - Invulnerable</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<h1 class="text-3xl font-bold text-gray-900">Software Bill of Materials</h1>
		<a href="/scans/{scanId}" class="text-blue-600 hover:text-blue-800">← Back to Scan</a>
	</div>

	{#if loading}
		<div class="text-center py-12">
			<p class="text-gray-500">Loading SBOM...</p>
		</div>
	{:else if error}
		<div class="card bg-red-50">
			<p class="text-red-600">{error}</p>
		</div>
	{:else if sbom}
		<div class="card">
			<div class="flex justify-between items-center mb-4">
				<h2 class="text-xl font-semibold text-gray-900">Scan #{scanId}</h2>
				<div class="space-x-2">
					<button on:click={copySBOM} class="btn btn-secondary">
						{copied ? 'Copied!' : 'Copy to Clipboard'}
					</button>
					<button on:click={downloadSBOM} class="btn btn-primary">Download JSON</button>
				</div>
			</div>

			<div class="bg-gray-900 rounded-lg p-4 overflow-auto max-h-[600px]">
				<pre class="text-sm text-gray-100"><code>{sbomJson}</code></pre>
			</div>
		</div>

		{#if sbom.artifacts && Array.isArray(sbom.artifacts)}
			<div class="card">
				<h2 class="text-xl font-semibold text-gray-900 mb-4">
					Package Summary ({sbom.artifacts.length} packages)
				</h2>
				<div class="overflow-x-auto max-h-[400px]">
					<table class="min-w-full divide-y divide-gray-200">
						<thead class="bg-gray-50 sticky top-0">
							<tr>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Version</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
							</tr>
						</thead>
						<tbody class="bg-white divide-y divide-gray-200">
							{#each sbom.artifacts as artifact}
								<tr class="hover:bg-gray-50">
									<td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
										{artifact.name}
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
										{artifact.version || 'N/A'}
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
										{artifact.type || 'N/A'}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		{/if}
	{/if}
</div>

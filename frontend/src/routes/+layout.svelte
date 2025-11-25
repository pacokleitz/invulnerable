<script lang="ts">
	import '../app.css';
</script>

<div class="min-h-screen bg-gray-50">
	<!-- Navigation -->
	<nav class="bg-white shadow-sm border-b border-gray-200">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<div class="flex justify-between h-16">
				<div class="flex">
					<div class="flex-shrink-0 flex items-center">
						<a href="/" class="text-xl font-bold text-blue-600">Invulnerable</a>
					</div>
					<div class="ml-6 flex space-x-8">
						<a
							href="/"
							class="inline-flex items-center px-1 pt-1 text-sm font-medium text-gray-900 hover:text-blue-600"
						>
							Dashboard
						</a>
						<a
							href="/scans"
							class="inline-flex items-center px-1 pt-1 text-sm font-medium text-gray-500 hover:text-blue-600"
						>
							Scans
						</a>
						<a
							href="/vulnerabilities"
							class="inline-flex items-center px-1 pt-1 text-sm font-medium text-gray-500 hover:text-blue-600"
						>
							Vulnerabilities
						</a>
						<a
							href="/images"
							class="inline-flex items-center px-1 pt-1 text-sm font-medium text-gray-500 hover:text-blue-600"
						>
							Images
						</a>
					</div>
				</div>
			</div>
		</div>
	</nav>

	<!-- Main content -->
	<main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
		<slot />
	</main>
</div>

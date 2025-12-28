// Package type categorization for CVE ownership routing

export type PackageCategory = 'os' | 'application' | 'unknown';

export interface PackageCategoryInfo {
	category: PackageCategory;
	label: string;
	description: string;
	color: string;
	bgColor: string;
	owner: string;
}

// Map of package types from Syft/Grype to their categories
const PACKAGE_TYPE_CATEGORIES: Record<string, PackageCategory> = {
	// OS/Filesystem packages - managed by infrastructure/DevOps
	'apk': 'os',           // Alpine packages
	'deb': 'os',           // Debian/Ubuntu packages
	'rpm': 'os',           // RedHat/CentOS/Fedora packages
	'portage': 'os',       // Gentoo packages
	'alpm': 'os',          // Arch Linux packages
	'dpkg': 'os',          // Debian package manager

	// Application/Library packages - managed by developers
	'npm': 'application',           // Node.js packages
	'yarn': 'application',          // Node.js (Yarn)
	'pnpm': 'application',          // Node.js (pnpm)
	'pip': 'application',           // Python packages
	'python': 'application',        // Python packages
	'gem': 'application',           // Ruby gems
	'bundler': 'application',       // Ruby (Bundler)
	'cargo': 'application',         // Rust crates
	'go-module': 'application',     // Go modules
	'gomod': 'application',         // Go modules
	'java-archive': 'application',  // Java JAR/WAR
	'jar': 'application',           // Java
	'maven': 'application',         // Java (Maven)
	'gradle': 'application',        // Java (Gradle)
	'nuget': 'application',         // .NET packages
	'composer': 'application',      // PHP packages
	'cocoapods': 'application',     // iOS/macOS packages
	'swift': 'application',         // Swift packages
	'pub': 'application',           // Dart/Flutter packages
	'hex': 'application',           // Erlang/Elixir packages
	'hackage': 'application',       // Haskell packages
};

export function categorizePackageType(packageType?: string | null): PackageCategoryInfo {
	if (!packageType) {
		return {
			category: 'unknown',
			label: 'Unknown',
			description: 'Package type not specified',
			color: 'text-gray-700',
			bgColor: 'bg-gray-100',
			owner: 'Unknown',
		};
	}

	const normalizedType = packageType.toLowerCase();
	const category = PACKAGE_TYPE_CATEGORIES[normalizedType] || 'unknown';

	switch (category) {
		case 'os':
			return {
				category: 'os',
				label: 'OS Package',
				description: 'Operating system or filesystem package',
				color: 'text-purple-700',
				bgColor: 'bg-purple-100',
				owner: 'DevOps',
			};
		case 'application':
			return {
				category: 'application',
				label: 'Application',
				description: 'Application dependency or library',
				color: 'text-blue-700',
				bgColor: 'bg-blue-100',
				owner: 'Developers',
			};
		default:
			return {
				category: 'unknown',
				label: packageType,
				description: `Unknown package type: ${packageType}`,
				color: 'text-gray-700',
				bgColor: 'bg-gray-100',
				owner: 'Unknown',
			};
	}
}

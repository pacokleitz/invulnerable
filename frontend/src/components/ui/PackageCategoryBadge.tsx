import { FC } from 'react';
import { categorizePackageType } from '../../lib/utils/packageTypes';

interface PackageCategoryBadgeProps {
	packageType?: string | null;
}

export const PackageCategoryBadge: FC<PackageCategoryBadgeProps> = ({ packageType }) => {
	const info = categorizePackageType(packageType);

	return (
		<span
			className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${info.bgColor} ${info.color}`}
			title={info.description}
		>
			{info.label}
		</span>
	);
};

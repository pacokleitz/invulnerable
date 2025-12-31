import { FC, useState } from 'react';

export type SortDirection = 'asc' | 'desc' | null;

interface SortableTableHeaderProps {
	label: string;
	sortKey: string;
	currentSortKey: string | null;
	currentSortDirection: SortDirection;
	onSort: (key: string) => void;
	className?: string;
}

export const SortableTableHeader: FC<SortableTableHeaderProps> = ({
	label,
	sortKey,
	currentSortKey,
	currentSortDirection,
	onSort,
	className = '',
}) => {
	const isActive = currentSortKey === sortKey;

	return (
		<th
			className={`px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 select-none ${className}`}
			onClick={() => onSort(sortKey)}
		>
			<div className="flex items-center space-x-1">
				<span>{label}</span>
				{isActive && (
					<span className="text-gray-400">
						{currentSortDirection === 'asc' ? '↑' : '↓'}
					</span>
				)}
			</div>
		</th>
	);
};

// Helper hook for managing sort state
export const useSortState = (defaultKey?: string, defaultDirection: SortDirection = 'desc') => {
	const [sortKey, setSortKey] = useState<string | null>(defaultKey || null);
	const [sortDirection, setSortDirection] = useState<SortDirection>(defaultDirection);

	const handleSort = (key: string) => {
		if (sortKey === key) {
			// Toggle direction or clear sort
			if (sortDirection === 'desc') {
				setSortDirection('asc');
			} else if (sortDirection === 'asc') {
				setSortDirection(null);
				setSortKey(null);
			}
		} else {
			// New column, start with desc
			setSortKey(key);
			setSortDirection('desc');
		}
	};

	return {
		sortKey,
		sortDirection,
		handleSort,
	};
};

import { FC } from 'react';

interface StatusBadgeProps {
	status: string;
	onClick?: () => void;
	title?: string;
}

export const StatusBadge: FC<StatusBadgeProps> = ({ status, onClick, title }) => {
	const badgeClass =
		{
			active: 'badge-active',
			in_progress: 'badge-in-progress',
			fixed: 'badge-fixed',
			ignored: 'badge-ignored',
			accepted: 'badge-accepted'
		}[status] || 'badge-active';

	if (onClick) {
		return (
			<button
				onClick={onClick}
				title={title || 'View change history'}
				className={`badge ${badgeClass} cursor-pointer hover:opacity-80 transition-opacity`}
			>
				{status}
			</button>
		);
	}

	return <span className={`badge ${badgeClass}`}>{status}</span>;
};

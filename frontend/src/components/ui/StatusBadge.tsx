import { FC } from 'react';

interface StatusBadgeProps {
	status: string;
}

export const StatusBadge: FC<StatusBadgeProps> = ({ status }) => {
	const badgeClass =
		{
			active: 'badge-active',
			fixed: 'badge-fixed',
			ignored: 'badge-ignored',
			accepted: 'badge-accepted'
		}[status] || 'badge-active';

	return <span className={`badge ${badgeClass}`}>{status}</span>;
};

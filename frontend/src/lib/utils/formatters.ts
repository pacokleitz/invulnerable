export const formatDate = (date: string): string => {
	return new Date(date).toLocaleString();
};

export const truncateText = (text: string, maxLength: number): string => {
	if (text.length <= maxLength) return text;
	return `${text.substring(0, maxLength)}...`;
};

export const daysSince = (date: string): number => {
	const then = new Date(date);
	const now = new Date();
	const diffMs = now.getTime() - then.getTime();
	const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
	return diffDays;
};

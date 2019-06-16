#include <stdio.h>
 
int get_sum(int n)
{
    int sum = 0, i;
    for (i = 1; i <= n; i++)
        sum += i;
 
    return sum;
}
 
int main()
{
    int input;
 
    while (scanf("%d", &input) != EOF)
        printf("%d\n\n", get_sum(input));
 
    return 0;
}

package templates

func ChildCommandHelpFunc(cmd *cobra.Command, args []string) {
	// 打印子命令自己的 Flag 信息
	fmt.Println("Usage:")
	fmt.Println(cmd.UseLine())
	fmt.Println("\nFlags:")
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		fmt.Printf("  --%s: %s\n", f.Name, f.Usage)
	})

	// 打印继承自父命令的 Persistent Flags 信息
	fmt.Println("\nInherited Flags (from parent):")
	cmd.InheritedFlags().VisitAll(func(f *pflag.Flag) {
		fmt.Printf("  --%s: %s\n", f.Name, f.Usage)
	})
}
